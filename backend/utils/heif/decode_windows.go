//go:build windows && cgo

package heif

/*
#include <windows.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Minimal subset of the libheif C ABI, declared here instead of including
// libheif's headers so EMLy compiles without them: libheif.dll is only
// needed at runtime. libheif keeps this ABI stable across releases; values
// are from heif_error.h / heif_image.h (libheif 1.23).
typedef struct {
	int code;              // enum heif_error_code, 0 = heif_error_Ok
	int subcode;           // enum heif_suberror_code
	const char* message;
} emly_heif_error;

enum {
	EMLY_HEIF_ERR_USAGE = 5,
	EMLY_HEIF_ERR_MEMORY = 6,
	EMLY_HEIF_ERR_DECODER = 7,
	EMLY_HEIF_COLORSPACE_RGB = 1,
	EMLY_HEIF_CHROMA_INTERLEAVED_RGBA = 11,
	EMLY_HEIF_CHANNEL_INTERLEAVED = 10,
};

typedef emly_heif_error (*fn_init)(void* params);
typedef const char* (*fn_get_version)(void);
typedef void* (*fn_context_alloc)(void);
typedef void (*fn_context_free)(void* ctx);
typedef emly_heif_error (*fn_read_from_memory)(void* ctx, const void* mem, size_t size, const void* options);
typedef emly_heif_error (*fn_get_primary)(void* ctx, void** handle);
typedef emly_heif_error (*fn_decode_image)(const void* handle, void** img, int colorspace, int chroma, const void* options);
typedef int (*fn_image_get_dim)(const void* img, int channel);
typedef const uint8_t* (*fn_plane_readonly2)(const void* img, int channel, size_t* stride);
typedef const uint8_t* (*fn_plane_readonly)(const void* img, int channel, int* stride);
typedef void (*fn_release)(const void* obj);

static struct {
	fn_get_version get_version;
	fn_context_alloc context_alloc;
	fn_context_free context_free;
	fn_read_from_memory read_from_memory;
	fn_get_primary get_primary;
	fn_decode_image decode_image;
	fn_image_get_dim image_get_width;
	fn_image_get_dim image_get_height;
	fn_plane_readonly2 plane_readonly2;  // libheif >= 1.20, preferred
	fn_plane_readonly plane_readonly;    // older libheif fallback
	fn_release image_release;
	fn_release handle_release;
} lh;

// emly_heif_load loads libheif.dll from the absolute path `path` and
// resolves the functions EMLy uses. LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR |
// SYSTEM32 restricts the DLL's own dependencies to its folder and System32,
// so nothing is picked up from the current directory or PATH. Returns 0 on
// success, otherwise writes a reason to err.
static int emly_heif_load(const wchar_t* path, char* err, size_t err_len) {
	HMODULE dll = LoadLibraryExW(path, NULL,
		LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR | LOAD_LIBRARY_SEARCH_SYSTEM32);
	if (!dll) {
		snprintf(err, err_len, "LoadLibraryExW failed (Win32 error %lu)", GetLastError());
		return -1;
	}

#define EMLY_RESOLVE(field, type, name) \
	lh.field = (type)(void*)GetProcAddress(dll, name); \
	if (!lh.field) { snprintf(err, err_len, "missing export %s", name); FreeLibrary(dll); return -1; }

	fn_init init;
	EMLY_RESOLVE(get_version, fn_get_version, "heif_get_version");
	EMLY_RESOLVE(context_alloc, fn_context_alloc, "heif_context_alloc");
	EMLY_RESOLVE(context_free, fn_context_free, "heif_context_free");
	EMLY_RESOLVE(read_from_memory, fn_read_from_memory, "heif_context_read_from_memory_without_copy");
	EMLY_RESOLVE(get_primary, fn_get_primary, "heif_context_get_primary_image_handle");
	EMLY_RESOLVE(decode_image, fn_decode_image, "heif_decode_image");
	EMLY_RESOLVE(image_get_width, fn_image_get_dim, "heif_image_get_width");
	EMLY_RESOLVE(image_get_height, fn_image_get_dim, "heif_image_get_height");
	EMLY_RESOLVE(image_release, fn_release, "heif_image_release");
	EMLY_RESOLVE(handle_release, fn_release, "heif_image_handle_release");
#undef EMLY_RESOLVE

	init = (fn_init)(void*)GetProcAddress(dll, "heif_init");
	lh.plane_readonly2 = (fn_plane_readonly2)(void*)GetProcAddress(dll, "heif_image_get_plane_readonly2");
	lh.plane_readonly = (fn_plane_readonly)(void*)GetProcAddress(dll, "heif_image_get_plane_readonly");
	if (!lh.plane_readonly2 && !lh.plane_readonly) {
		snprintf(err, err_len, "missing export heif_image_get_plane_readonly");
		FreeLibrary(dll);
		return -1;
	}

	// heif_init is optional (pre-1.13 libheif lacks it); never deinit, the
	// DLL stays loaded for the life of the process.
	if (init) {
		emly_heif_error e = init(NULL);
		if (e.code != 0) {
			snprintf(err, err_len, "heif_init: %s", e.message ? e.message : "?");
			FreeLibrary(dll);
			return -1;
		}
	}
	return 0;
}

static const char* emly_heif_version(void) {
	return lh.get_version ? lh.get_version() : "";
}

typedef struct {
	int width;
	int height;
	int code;          // heif_error_code; 0 = ok
	int subcode;
	char message[256];
} emly_heif_result;

static void emly_set_err(emly_heif_result* r, int code, int subcode, const char* msg) {
	r->code = code;
	r->subcode = subcode;
	strncpy(r->message, msg ? msg : "", sizeof(r->message) - 1);
	r->message[sizeof(r->message) - 1] = 0;
}

// emly_heif_decode decodes the primary image of an in-memory HEIF file into
// interleaved 8-bit RGBA and returns a malloc'd, tightly packed copy of the
// pixels (w*h*4 bytes; caller frees), or NULL with r->code set. Every
// libheif object lives and dies inside this call. Requires a successful
// emly_heif_load.
static uint8_t* emly_heif_decode(const void* data, size_t size, emly_heif_result* r) {
	memset(r, 0, sizeof(*r));
	uint8_t* dst = NULL;
	void* handle = NULL;
	void* img = NULL;
	emly_heif_error e;

	void* ctx = lh.context_alloc();
	if (!ctx) {
		emly_set_err(r, EMLY_HEIF_ERR_MEMORY, 0, "heif_context_alloc failed");
		return NULL;
	}

	e = lh.read_from_memory(ctx, data, size, NULL);
	if (e.code != 0) { emly_set_err(r, e.code, e.subcode, e.message); goto done; }

	e = lh.get_primary(ctx, &handle);
	if (e.code != 0) { emly_set_err(r, e.code, e.subcode, e.message); goto done; }

	// Default decoding options apply the irot/imir transformations, so the
	// photo comes out upright.
	e = lh.decode_image(handle, &img, EMLY_HEIF_COLORSPACE_RGB, EMLY_HEIF_CHROMA_INTERLEAVED_RGBA, NULL);
	if (e.code != 0) { emly_set_err(r, e.code, e.subcode, e.message); goto done; }

	r->width = lh.image_get_width(img, EMLY_HEIF_CHANNEL_INTERLEAVED);
	r->height = lh.image_get_height(img, EMLY_HEIF_CHANNEL_INTERLEAVED);
	if (r->width <= 0 || r->height <= 0) {
		emly_set_err(r, EMLY_HEIF_ERR_DECODER, 0, "decoded image has no interleaved plane");
		goto done;
	}

	size_t stride = 0;
	const uint8_t* plane;
	if (lh.plane_readonly2) {
		plane = lh.plane_readonly2(img, EMLY_HEIF_CHANNEL_INTERLEAVED, &stride);
	} else {
		int istride = 0;
		plane = lh.plane_readonly(img, EMLY_HEIF_CHANNEL_INTERLEAVED, &istride);
		stride = istride > 0 ? (size_t)istride : 0;
	}
	size_t row = (size_t)r->width * 4;
	if (!plane || stride < row) {
		emly_set_err(r, EMLY_HEIF_ERR_DECODER, 0, "unexpected plane layout");
		goto done;
	}
	dst = (uint8_t*)malloc(row * (size_t)r->height);
	if (!dst) {
		emly_set_err(r, EMLY_HEIF_ERR_MEMORY, 0, "out of memory");
		goto done;
	}
	for (int y = 0; y < r->height; y++) {
		memcpy(dst + (size_t)y * row, plane + (size_t)y * stride, row);
	}

done:
	if (img) lh.image_release(img);
	if (handle) lh.handle_release(handle);
	lh.context_free(ctx);
	return dst;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

// DLLName is the file EMLy loads from its own folder.
const DLLName = "libheif.dll"

var (
	loadOnce    sync.Once
	loadErr     error
	libraryPath string // empty: <exe dir>\libheif.dll
	version     string
)

// SetLibraryPath overrides where libheif.dll is loaded from (tests, dev
// tools). It must be called before the first Available/Decode call and
// should only ever be given trusted, absolute paths.
func SetLibraryPath(path string) { libraryPath = path }

func defaultLibraryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), DLLName), nil
}

func load() error {
	loadOnce.Do(func() {
		path := libraryPath
		if path == "" {
			var err error
			if path, err = defaultLibraryPath(); err != nil {
				loadErr = fmt.Errorf("heif: locate %s: %w", DLLName, err)
				return
			}
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			loadErr = fmt.Errorf("heif: %w", err)
			return
		}
		if _, err := os.Stat(abs); err != nil {
			loadErr = fmt.Errorf("%w: %s not found", ErrUnavailable, abs)
			return
		}
		wpath, err := syscall.UTF16PtrFromString(abs)
		if err != nil {
			loadErr = fmt.Errorf("heif: %w", err)
			return
		}
		var cerr [256]C.char
		if C.emly_heif_load((*C.wchar_t)(unsafe.Pointer(wpath)), &cerr[0], C.size_t(len(cerr))) != 0 {
			loadErr = fmt.Errorf("%w: %s: %s", ErrUnavailable, abs, C.GoString(&cerr[0]))
			return
		}
		version = C.GoString(C.emly_heif_version())
	})
	return loadErr
}

func resultErr(r *C.emly_heif_result) error {
	if r.code == 0 {
		return nil
	}
	return &Error{Code: int(r.code), Subcode: int(r.subcode), Message: C.GoString(&r.message[0])}
}

// Decode decodes the primary image of an HEIF file to NRGBA (libheif emits
// non-premultiplied alpha). libheif's default security limits (max image
// size, max item count, ...) apply, so untrusted attachments cannot make it
// allocate unbounded memory.
func Decode(data []byte) (*image.NRGBA, error) {
	if err := load(); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, &Error{Message: "empty input"}
	}

	// libheif keeps referencing the input across calls, which cgo pointer
	// rules forbid for Go memory: hand it a C copy.
	cdata := C.CBytes(data)
	defer C.free(cdata)

	var r C.emly_heif_result
	pix := C.emly_heif_decode(cdata, C.size_t(len(data)), &r)
	if pix == nil {
		if err := resultErr(&r); err != nil {
			return nil, err
		}
		return nil, errors.New("heif: decode failed")
	}
	defer C.free(unsafe.Pointer(pix))

	img := image.NewNRGBA(image.Rect(0, 0, int(r.width), int(r.height)))
	copy(img.Pix, unsafe.Slice((*byte)(unsafe.Pointer(pix)), len(img.Pix)))
	return img, nil
}
