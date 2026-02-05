/**
 * Utility functions for mail data conversion and processing
 */

/**
 * Converts an ArrayBuffer or byte array to a base64 string
 * @param buffer - The buffer to convert (can be string, array, or ArrayBuffer)
 * @returns Base64 encoded string
 */
export function arrayBufferToBase64(buffer: unknown): string {
  // Already a base64 string
  if (typeof buffer === 'string') {
    return buffer;
  }

  // Handle array of bytes
  if (Array.isArray(buffer)) {
    let binary = '';
    const bytes = new Uint8Array(buffer);
    const len = bytes.byteLength;
    for (let i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary);
  }

  // Handle ArrayBuffer
  if (buffer instanceof ArrayBuffer) {
    let binary = '';
    const bytes = new Uint8Array(buffer);
    const len = bytes.byteLength;
    for (let i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary);
  }

  return '';
}

/**
 * Creates a data URL for downloading attachments
 * @param contentType - MIME type of the attachment
 * @param base64Data - Base64 encoded data
 * @returns Data URL string
 */
export function createDataUrl(contentType: string, base64Data: string): string {
  return `data:${contentType};base64,${base64Data}`;
}

/**
 * Checks if a string looks like valid base64 content
 * @param content - String to check
 * @returns True if the content appears to be base64 encoded
 */
export function looksLikeBase64(content: string): boolean {
  const clean = content.replace(/[\s\r\n]+/g, '');
  return (
    clean.length > 0 &&
    clean.length % 4 === 0 &&
    /^[A-Za-z0-9+/]+=*$/.test(clean)
  );
}

/**
 * Attempts to decode base64 content
 * @param content - Base64 string to decode
 * @returns Decoded string or null if decoding fails
 */
export function tryDecodeBase64(content: string): string | null {
  try {
    const clean = content.replace(/[\s\r\n]+/g, '');
    return window.atob(clean);
  } catch {
    return null;
  }
}
