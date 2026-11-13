/**
 * Client-side image -> displayable-image conversion.
 *
 * EMLy's built-in image viewers receive either base64 bytes from the Go
 * backend (viewer windows: app_viewer.go GetImageViewerData/GetViewerData,
 * see toDisplayableImageSrc) or a Blob of raw bytes from the AssetServer
 * (image tabs: app_attachments.go, see toDisplayableImageFromBlob). Browsers/WebView2 cannot
 * natively decode HEIC/HEIF or TIFF for an <img> tag, so for those formats we
 * decode to a JPEG in-browser - HEIC via libheif-js (WASM), TIFF via utif -
 * purely for on-screen display.
 *
 * HEIC/HEIF is normally converted to JPEG by the Go backend already (native
 * libheif.dll, backend/utils/heif): such bytes sniff as JPEG behind a .heic
 * name (see isBackendConvertedHeic). libheif-js stays as the fallback when
 * libheif.dll is missing, for preloaded base64 attachments and for files the
 * backend failed to decode.
 *
 * The format is detected from the file's magic bytes rather than trusted from
 * the extension: scanners and document archives commonly emit TIFF files
 * named ".jpeg"/".jpg", which would otherwise show up as a broken image.
 */

const HEIC_EXTENSIONS = [".heic", ".heif"];
const TIFF_EXTENSIONS = [".tif", ".tiff"];

/** True if filename has a .heic/.heif extension (case-insensitive). */
export function isHeic(filename: string): boolean {
  const lower = filename.toLowerCase();
  return HEIC_EXTENSIONS.some((ext) => lower.endsWith(ext));
}

function isTiffFilename(filename: string): boolean {
  const lower = filename.toLowerCase();
  return TIFF_EXTENSIONS.some((ext) => lower.endsWith(ext));
}

const MIME_BY_EXT: Record<string, string> = {
  jpg: "image/jpeg",
  jpeg: "image/jpeg",
  png: "image/png",
  gif: "image/gif",
  bmp: "image/bmp",
  webp: "image/webp",
  tif: "image/tiff",
  tiff: "image/tiff",
};

/** Best-effort MIME type for a filename's extension, defaulting to image/png. */
export function mimeTypeForFilename(filename: string): string {
  const ext = filename.split(".").pop()?.toLowerCase() ?? "";
  return MIME_BY_EXT[ext] ?? "image/png";
}

type SniffedFormat = "jpeg" | "png" | "gif" | "bmp" | "webp" | "tiff" | "heic";

const HEIC_BRANDS = new Set(["heic", "heix", "hevc", "hevx", "heim", "heis", "mif1", "msf1"]);

/** Identifies the image format from its leading bytes; null if unknown. */
function sniffImageFormat(bytes: Uint8Array): SniffedFormat | null {
  const ascii = (start: number, end: number) =>
    String.fromCharCode(...bytes.subarray(start, end));

  if (bytes[0] === 0xff && bytes[1] === 0xd8 && bytes[2] === 0xff) return "jpeg";
  if (bytes[0] === 0x89 && ascii(1, 4) === "PNG") return "png";
  if (ascii(0, 4) === "GIF8") return "gif";
  if (ascii(0, 2) === "BM") return "bmp";
  if (ascii(0, 4) === "RIFF" && ascii(8, 12) === "WEBP") return "webp";
  if (
    (bytes[0] === 0x49 && bytes[1] === 0x49 && bytes[2] === 0x2a && bytes[3] === 0x00) ||
    (bytes[0] === 0x4d && bytes[1] === 0x4d && bytes[2] === 0x00 && bytes[3] === 0x2a)
  ) {
    return "tiff";
  }
  if (ascii(4, 8) === "ftyp" && HEIC_BRANDS.has(ascii(8, 12))) return "heic";
  return null;
}

/** Sniffs the format from the first bytes of a Blob (reads only 16 bytes). */
async function sniffBlob(blob: Blob): Promise<SniffedFormat | null> {
  try {
    return sniffImageFormat(new Uint8Array(await blob.slice(0, 16).arrayBuffer()));
  } catch {
    return null;
  }
}

/** Decodes just enough of the base64 payload to sniff its magic bytes. */
function sniffBase64(base64Data: string): SniffedFormat | null {
  try {
    // 16 base64 chars = 12 bytes, enough for every signature above.
    return sniffImageFormat(base64ToUint8Array(base64Data.slice(0, 16)));
  } catch {
    return null;
  }
}

export interface DisplayableImage {
  /** A ready-to-use <img src> value: a data: URI, for both the fast path
   *  and the decoded JPEG. */
  src: string;
  /** Present only when the download button should offer different bytes
   *  than the original base64Data/filename: HEIC/HEIF, or TIFF data hidden
   *  behind a non-TIFF extension - converted to JPEG since the original
   *  can't be opened as its name suggests. */
  download?: { base64: string; filename: string };
}

/**
 * Converts base64 image bytes into something an <img> tag can render.
 *
 * Browser-native formats: fast path, returns a plain `data:<mime>;base64,...`
 * URI, with the MIME type taken from the sniffed content when recognized.
 *
 * HEIC/HEIF and TIFF (by content, or by extension when the content is not
 * recognized): decoded to a JPEG via an offscreen canvas. The decoders are
 * dynamically imported so they only load when actually needed. Throws a
 * normal Error if decoding fails - callers must catch and show an error state.
 */
export async function toDisplayableImageSrc(
  base64Data: string,
  filename: string,
  mimeType: string,
): Promise<DisplayableImage> {
  const sniffed = sniffBase64(base64Data);
  const format = sniffed ?? (isHeic(filename) ? "heic" : isTiffFilename(filename) ? "tiff" : null);

  if (isBackendConvertedHeic(format, filename)) {
    return {
      src: `data:image/jpeg;base64,${base64Data}`,
      download: { base64: base64Data, filename: withJpgExtension(filename) },
    };
  }

  try {
    if (format === "heic") {
      const dataUrl = await decodeHeicToJpeg(base64Data);
      return {
        src: dataUrl,
        download: { base64: dataUrl.split(",")[1] ?? "", filename: withJpgExtension(filename) },
      };
    }

    if (format === "tiff") {
      const dataUrl = await decodeTiffToJpeg(base64Data);
      // A real .tif/.tiff opens fine in other apps, so keep the original
      // bytes for download; only swap in the JPEG when the name lies.
      return isTiffFilename(filename)
        ? { src: dataUrl }
        : {
            src: dataUrl,
            download: { base64: dataUrl.split(",")[1] ?? "", filename: withJpgExtension(filename) },
          };
    }
  } catch (err) {
    throw normalizeDecodeError(err);
  }

  const mime = format ? `image/${format}` : mimeType;
  return { src: `data:${mime};base64,${base64Data}` };
}

/**
 * Blob counterpart of toDisplayableImageSrc, for viewers that get raw bytes
 * (AssetServer) instead of base64. Browser-native formats become an object
 * URL of the Blob itself - no base64 copy, no data: URI; `objectUrl` is
 * true and the caller must URL.revokeObjectURL(src) when done. HEIC/HEIF and
 * TIFF are decoded to a JPEG data: URI exactly like toDisplayableImageSrc.
 */
export async function toDisplayableImageFromBlob(
  blob: Blob,
  filename: string,
  mimeType: string,
): Promise<DisplayableImage & { objectUrl: boolean }> {
  const sniffed = await sniffBlob(blob);
  const format = sniffed ?? (isHeic(filename) ? "heic" : isTiffFilename(filename) ? "tiff" : null);

  if (isBackendConvertedHeic(format, filename)) {
    const jpeg = new Blob([blob], { type: "image/jpeg" });
    return {
      src: URL.createObjectURL(jpeg),
      objectUrl: true,
      download: { base64: await blobToBase64(jpeg), filename: withJpgExtension(filename) },
    };
  }

  if (format === "heic" || format === "tiff") {
    const bytes = new Uint8Array(await blob.arrayBuffer());
    try {
      if (format === "heic") {
        const dataUrl = await decodeHeicBytesToJpeg(bytes);
        return {
          src: dataUrl,
          objectUrl: false,
          download: { base64: dataUrl.split(",")[1] ?? "", filename: withJpgExtension(filename) },
        };
      }
      const dataUrl = await decodeTiffBytesToJpeg(bytes);
      return isTiffFilename(filename)
        ? { src: dataUrl, objectUrl: false }
        : {
            src: dataUrl,
            objectUrl: false,
            download: { base64: dataUrl.split(",")[1] ?? "", filename: withJpgExtension(filename) },
          };
    } catch (err) {
      throw normalizeDecodeError(err);
    }
  }

  const mime = format ? `image/${format}` : mimeType;
  return { src: URL.createObjectURL(new Blob([blob], { type: mime })), objectUrl: true };
}

async function decodeHeicToJpeg(base64Data: string): Promise<string> {
  return decodeHeicBytesToJpeg(base64ToUint8Array(base64Data));
}

async function decodeHeicBytesToJpeg(bytes: Uint8Array): Promise<string> {
  const mod = await import("libheif-js/wasm-bundle");
  // The wasm-bundle package is CommonJS; depending on bundler interop the
  // named export may land on `.default` or directly on the module object.
  const libheif = mod.default ?? mod;
  const decoder = new libheif.HeifDecoder();
  const images = decoder.decode(bytes);
  if (!images || images.length === 0) {
    throw new Error("No image found in HEIC/HEIF file");
  }

  const image = images[0];
  const { canvas, ctx } = createCanvas(image.get_width(), image.get_height());
  const imageData = ctx.createImageData(canvas.width, canvas.height);

  await new Promise<void>((resolve, reject) => {
    image.display(imageData, (displayData) => {
      if (!displayData) {
        reject(new Error("Failed to render HEIC/HEIF image"));
        return;
      }
      resolve();
    });
  });
  ctx.putImageData(imageData, 0, 0);

  return canvas.toDataURL("image/jpeg", 0.92);
}

async function decodeTiffToJpeg(base64Data: string): Promise<string> {
  return decodeTiffBytesToJpeg(base64ToUint8Array(base64Data));
}

async function decodeTiffBytesToJpeg(bytes: Uint8Array): Promise<string> {
  const mod = await import("utif");
  // CommonJS package: same interop caveat as libheif-js above.
  const UTIF = mod.default ?? mod;
  const ifds = UTIF.decode(bytes);
  // Multi-page TIFFs show their first page; skip non-image directories.
  const page = ifds.find((ifd) => ifd["t256"] != null);
  if (!page) {
    throw new Error("No image found in TIFF file");
  }
  UTIF.decodeImage(bytes, page, ifds);
  if (!page.width || !page.height) {
    throw new Error("Failed to decode TIFF image");
  }

  const { canvas, ctx } = createCanvas(page.width, page.height);
  const imageData = ctx.createImageData(page.width, page.height);
  imageData.data.set(UTIF.toRGBA8(page));
  // Transparent TIFFs would turn black in JPEG; flatten onto white.
  ctx.fillStyle = "#fff";
  ctx.fillRect(0, 0, page.width, page.height);
  const layer = createCanvas(page.width, page.height);
  layer.ctx.putImageData(imageData, 0, 0);
  ctx.drawImage(layer.canvas, 0, 0);

  return canvas.toDataURL("image/jpeg", 0.92);
}

function createCanvas(width: number, height: number) {
  const canvas = document.createElement("canvas");
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext("2d");
  if (!ctx) {
    throw new Error("Canvas 2D context unavailable");
  }
  return { canvas, ctx };
}

/** libheif can throw plain {code, subcode} objects rather than Errors -
 *  normalize so callers can safely display `error.message`. */
function normalizeDecodeError(err: unknown): Error {
  if (err instanceof Error) return err;
  try {
    return new Error(JSON.stringify(err));
  } catch {
    return new Error(String(err));
  }
}

/** True for JPEG bytes behind a .heic/.heif name: the Go backend already
 *  converted the HEIC (see app_heif.go). The download should then offer the
 *  JPEG under a .jpg name, like the in-browser decoder does. */
function isBackendConvertedHeic(format: SniffedFormat | null, filename: string): boolean {
  return format === "jpeg" && isHeic(filename);
}

function blobToBase64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result).split(",")[1] ?? "");
    reader.onerror = () => reject(reader.error ?? new Error("FileReader failed"));
    reader.readAsDataURL(blob);
  });
}

function base64ToUint8Array(base64Data: string): Uint8Array {
  const byteChars = atob(base64Data);
  const bytes = new Uint8Array(byteChars.length);
  for (let i = 0; i < byteChars.length; i++) {
    bytes[i] = byteChars.charCodeAt(i);
  }
  return bytes;
}

function withJpgExtension(filename: string): string {
  const dot = filename.lastIndexOf(".");
  const base = dot > 0 ? filename.slice(0, dot) : filename;
  return `${base}.jpg`;
}
