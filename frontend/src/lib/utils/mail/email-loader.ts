/**
 * Email loading and processing utilities
 */

import {
  ReadEML,
  ReadMSG,
  ReadPEC,
  ReadAuto,
  DetectEmailFormat,
  ShowOpenFileDialog,
  SetCurrentMailFilePath,
  ConvertToUTF8,
} from '$lib/wailsjs/go/main/App';
import type { internal } from '$lib/wailsjs/go/models';
import { isBase64, isHtml } from '$lib/utils';
import { looksLikeBase64, tryDecodeBase64 } from './data-utils';

export interface LoadEmailResult {
  success: boolean;
  email?: internal.EmailData;
  filePath?: string;
  error?: string;
  cancelled?: boolean;
}

/**
 * Determines the email file type from the path extension (best-effort hint).
 * Use DetectEmailFormat (backend) for reliable format detection.
 */
export function getEmailFileType(filePath: string): 'eml' | 'msg' | null {
  const lowerPath = filePath.toLowerCase();
  if (lowerPath.endsWith('.eml')) return 'eml';
  if (lowerPath.endsWith('.msg')) return 'msg';
  return null;
}

/**
 * Checks if a file path looks like an email file by extension.
 * Returns true also for unknown extensions so the backend can attempt parsing.
 */
export function isEmailFile(filePath: string): boolean {
  return filePath.trim().length > 0;
}

/**
 * Loads an email from a file path.
 * Uses ReadAuto so the backend detects the format from the file's binary
 * content, regardless of extension. Falls back to the legacy per-format
 * readers only when the caller explicitly requests them.
 *
 * @param filePath - Path to the email file
 * @returns LoadEmailResult with the email data or error
 */
export async function loadEmailFromPath(filePath: string): Promise<LoadEmailResult> {
  if (!filePath?.trim()) {
    return { success: false, error: 'No file path provided.' };
  }

  try {
    // ReadAuto detects the format (EML/PEC/MSG) by magic bytes and dispatches
    // to the appropriate reader. This works for any extension, including
    // unconventional ones like winmail.dat or no extension at all.
    const email = await ReadAuto(filePath);

    // Process body if needed (decode base64)
    if (email?.body) {
      const trimmed = email.body.trim();
      if (looksLikeBase64(trimmed)) {
        const decoded = tryDecodeBase64(trimmed);
        if (decoded) {
          email.body = decoded;
        }
      }
    }

    return { success: true, email, filePath };
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to load email:', error);
    return { success: false, error: errorMessage };
  }
}

/**
 * Loads an email using the explicit per-format readers (legacy path).
 * Prefer loadEmailFromPath for new code.
 */
export async function loadEmailFromPathLegacy(filePath: string): Promise<LoadEmailResult> {
  const fileType = getEmailFileType(filePath);

  if (!fileType) {
    return {
      success: false,
      error: 'Invalid file type. Only .eml and .msg files are supported.',
    };
  }

  try {
    let email: internal.EmailData;

    if (fileType === 'msg') {
      email = await ReadMSG(filePath, true);
    } else {
      try {
        email = await ReadPEC(filePath);
      } catch {
        email = await ReadEML(filePath);
      }
    }

    if (email?.body) {
      const trimmed = email.body.trim();
      if (looksLikeBase64(trimmed)) {
        const decoded = tryDecodeBase64(trimmed);
        if (decoded) {
          email.body = decoded;
        }
      }
    }

    return { success: true, email, filePath };
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to load email:', error);
    return { success: false, error: errorMessage };
  }
}

/**
 * Opens a file dialog and loads the selected email
 * @returns LoadEmailResult with the email data or error
 */
export async function openAndLoadEmail(): Promise<LoadEmailResult> {
  try {
    const filePath = await ShowOpenFileDialog();

    if (!filePath || filePath.length === 0) {
      return { success: false, cancelled: true };
    }

    const result = await loadEmailFromPath(filePath);

    if (result.success && result.email) {
      // Track the current mail file path for bug reports
      await SetCurrentMailFilePath(filePath);
    }

    return result;
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to open email file:', error);
    return {
      success: false,
      error: errorMessage,
    };
  }
}

/**
 * Processes and fixes the email body content
 * - Decodes base64 content if needed
 * - Fixes character encoding issues
 * @param body - The raw email body
 * @returns Processed body content
 */
export async function processEmailBody(body: string): Promise<string> {
  if (!body) return body;

  let content = body;

  // 1. Try to decode if not HTML
  if (!isHtml(content)) {
    const clean = content.replace(/[\s\r\n]+/g, '');
    if (isBase64(clean)) {
      const decoded = tryDecodeBase64(clean);
      if (decoded) {
        content = decoded;
      }
    }
  }

  // 2. If it is HTML (original or decoded), try to fix encoding
  if (isHtml(content)) {
    try {
      const fixed = await ConvertToUTF8(content);
      if (fixed) {
        content = fixed;
      }
    } catch (e) {
      console.warn('Failed to fix encoding:', e);
    }
  }

  return content;
}
