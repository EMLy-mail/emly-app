/**
 * Attachment sources: how the frontend refers to an attachment's bytes.
 *
 * - `AttachmentRef` (default): the attachment is addressed by its mail file
 *   path and index. Actions that end in Go (viewer windows, external apps,
 *   save to disk) pass the ref to a by-reference backend method, so the
 *   bytes never cross the Wails IPC bridge. When the frontend itself needs
 *   the bytes, it fetches them as raw binary from the AssetServer endpoint
 *   (see app_attachments.go) instead of a base64 IPC response.
 * - `string`: a base64 payload already in hand - "Old Pre-loading of
 *   attachments" (Settings → Danger Zone) or a viewer that only has base64.
 */

export interface AttachmentRef {
  /** Path of the mail file the attachment belongs to. */
  filePath: string;
  /** Index in the Attachments list returned by ReadAuto/ReadEML/... */
  index: number;
}

export type AttachmentSource = string | AttachmentRef;

/** AssetServer path serving raw attachment bytes (see app_attachments.go). */
const ATTACHMENT_ENDPOINT = '/emly-attachment';

export function isAttachmentRef(source: AttachmentSource): source is AttachmentRef {
  return typeof source !== 'string';
}

export interface AttachmentFetchOptions {
  /** The bytes are for an in-WebView viewer: lets the backend swap formats
   *  the WebView can't render (HEIC/HEIF) for a JPEG. Never set it when the
   *  bytes are saved or handed to another app. */
  display?: boolean;
}

/** URL of the raw attachment bytes for ref, served by the Go AssetServer. */
export function attachmentUrl(ref: AttachmentRef, opts: AttachmentFetchOptions = {}): string {
  const params = new URLSearchParams({ path: ref.filePath, index: String(ref.index) });
  if (opts.display) params.set('display', '1');
  return `${ATTACHMENT_ENDPOINT}?${params.toString()}`;
}

/** Fetches the raw attachment bytes for ref as a Blob. */
export async function fetchAttachmentBlob(
  ref: AttachmentRef,
  opts: AttachmentFetchOptions = {},
): Promise<Blob> {
  const res = await fetch(attachmentUrl(ref, opts), { cache: 'no-store' });
  if (!res.ok) {
    throw new Error(`attachment fetch failed: ${res.status} ${await res.text()}`);
  }
  return res.blob();
}

/** Decodes a base64 payload into bytes. */
export function base64ToBytes(base64Data: string): Uint8Array<ArrayBuffer> {
  const binary = atob(base64Data);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes;
}

/**
 * Resolves source to a Blob for in-WebView viewers (PDF/image tabs). A ref
 * is fetched as raw binary from the AssetServer endpoint, so no base64
 * string of the attachment ever exists in the renderer; a base64 source
 * (preload mode) is decoded once. type, when given, overrides the Blob's
 * MIME type (wrapping a Blob does not copy its bytes).
 */
export async function attachmentSourceToBlob(
  source: AttachmentSource,
  type?: string,
  opts: AttachmentFetchOptions = {},
): Promise<Blob> {
  const blob = isAttachmentRef(source)
    ? await fetchAttachmentBlob(source, opts)
    : new Blob([base64ToBytes(source)]);
  return type ? new Blob([blob], { type }) : blob;
}
