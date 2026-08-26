/**
 * Client-side HEIC/HEIF -> displayable-image conversion.
 *
 * EMLy's built-in image viewers only ever receive base64 bytes from the Go
 * backend, which just passes raw file bytes through unchanged (see
 * app_viewer.go GetImageViewerData/GetViewerData). Browsers/WebView2 cannot
 * natively decode HEIC/HEIF for an <img> tag, so for that one format we
 * decode to a JPEG in-browser via libheif-js (WASM), purely for on-screen
 * display - and, since most systems can't open a raw .heic file either, the
 * download button offers that same JPEG instead of the original bytes (see
 * `download` below).
 */

const HEIC_EXTENSIONS = [".heic", ".heif"];

/** True if filename has a .heic/.heif extension (case-insensitive). */
export function isHeic(filename: string): boolean {
  const lower = filename.toLowerCase();
  return HEIC_EXTENSIONS.some((ext) => lower.endsWith(ext));
}

const MIME_BY_EXT: Record<string, string> = {
  jpg: "image/jpeg",
  jpeg: "image/jpeg",
  png: "image/png",
  gif: "image/gif",
  bmp: "image/bmp",
  webp: "image/webp",
  tiff: "image/tiff",
};

/** Best-effort MIME type for a filename's extension, defaulting to image/png. */
export function mimeTypeForFilename(filename: string): string {
  const ext = filename.split(".").pop()?.toLowerCase() ?? "";
  return MIME_BY_EXT[ext] ?? "image/png";
}

export interface DisplayableImage {
  /** A ready-to-use <img src> value: a data: URI, for both the fast path
   *  (non-HEIC) and the HEIC-decoded JPEG. */
  src: string;
  /** Present only when the download button should offer different bytes
   *  than the original base64Data/filename (currently: HEIC/HEIF, converted
   *  to JPEG since most systems can't open the original format directly). */
  download?: { base64: string; filename: string };
}

/**
 * Converts base64 image bytes into something an <img> tag can render.
 *
 * Non-HEIC files: fast path, returns a plain `data:<mime>;base64,...` URI -
 * same behavior/perf as before this module existed.
 *
 * HEIC/HEIF files: decodes via libheif-js (dynamically imported so the WASM
 * payload only loads when actually needed) to a JPEG via an offscreen
 * canvas. Throws a normal Error if decoding fails - callers must catch and
 * show an error state.
 */
export async function toDisplayableImageSrc(
  base64Data: string,
  filename: string,
  mimeType: string,
): Promise<DisplayableImage> {
  if (!isHeic(filename)) {
    return { src: `data:${mimeType};base64,${base64Data}` };
  }

  try {
    const mod = await import("libheif-js/wasm-bundle");
    // The wasm-bundle package is CommonJS; depending on bundler interop the
    // named export may land on `.default` or directly on the module object.
    const libheif = mod.default ?? mod;
    const decoder = new libheif.HeifDecoder();
    const images = decoder.decode(base64ToUint8Array(base64Data));
    if (!images || images.length === 0) {
      throw new Error("No image found in HEIC/HEIF file");
    }

    const image = images[0];
    const width = image.get_width();
    const height = image.get_height();

    const canvas = document.createElement("canvas");
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      throw new Error("Canvas 2D context unavailable");
    }
    const imageData = ctx.createImageData(width, height);

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

    const dataUrl = canvas.toDataURL("image/jpeg", 0.92);
    const jpegBase64 = dataUrl.split(",")[1] ?? "";

    return {
      src: dataUrl,
      download: { base64: jpegBase64, filename: withJpgExtension(filename) },
    };
  } catch (err) {
    throw normalizeDecodeError(err);
  }
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
