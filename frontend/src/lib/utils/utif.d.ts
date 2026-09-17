/**
 * utif (UTIF.js) ships no TypeScript declarations. We only use its decode
 * pipeline (see image-decode.ts), so this stays a minimal ambient
 * declaration rather than a full API surface.
 */
declare module "utif" {
  /** An Image File Directory: raw TIFF tags keyed as "t<tag number>",
   *  plus width/height/data filled in by decodeImage. */
  interface IFD {
    [tag: string]: unknown;
    width: number;
    height: number;
    data: Uint8Array;
  }

  function decode(buffer: ArrayBuffer | Uint8Array): IFD[];
  function decodeImage(buffer: ArrayBuffer | Uint8Array, ifd: IFD, ifds?: IFD[]): void;
  function toRGBA8(ifd: IFD): Uint8Array;

  const _default: {
    decode: typeof decode;
    decodeImage: typeof decodeImage;
    toRGBA8: typeof toRGBA8;
  };
  export default _default;
  export { decode, decodeImage, toRGBA8 };
}
