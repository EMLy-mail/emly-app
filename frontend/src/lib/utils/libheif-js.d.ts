/**
 * libheif-js ships no TypeScript declarations for its "wasm-bundle" entry
 * point (a self-contained CommonJS build for browser/bundler use). We only
 * use a couple of its methods (see image-decode.ts), so this stays a
 * minimal ambient declaration rather than a full API surface.
 */
declare module "libheif-js/wasm-bundle" {
  interface HeifImage {
    get_width(): number;
    get_height(): number;
    display(
      imageData: ImageData,
      callback: (displayData: ImageData | null) => void,
    ): void;
  }

  class HeifDecoder {
    decode(bytes: Uint8Array): HeifImage[];
  }

  const _default: { HeifDecoder: typeof HeifDecoder };
  export default _default;
  export { HeifDecoder };
}
