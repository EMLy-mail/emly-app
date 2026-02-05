/**
 * Handlers for opening different attachment types
 */

import {
  OpenPDF,
  OpenPDFWindow,
  OpenImage,
  OpenImageWindow,
  OpenEMLWindow,
} from '$lib/wailsjs/go/main/App';
import { settingsStore } from '$lib/stores/settings.svelte';
import { toast } from 'svelte-sonner';
import * as m from '$lib/paraglide/messages';

export interface AttachmentHandlerResult {
  success: boolean;
  error?: string;
}

/**
 * Opens a PDF attachment using either built-in or external viewer
 * @param base64Data - Base64 encoded PDF data
 * @param filename - Name of the PDF file
 */
export async function openPDFAttachment(
  base64Data: string,
  filename: string
): Promise<AttachmentHandlerResult> {
  try {
    if (settingsStore.settings.useBuiltinPDFViewer) {
      await OpenPDFWindow(base64Data, filename);
    } else {
      await OpenPDF(base64Data, filename);
    }
    return { success: true };
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);

    // Check if PDF is already open
    if (errorMessage.includes(filename) && errorMessage.includes('already open')) {
      toast.error(m.mail_pdf_already_open());
      return { success: false, error: 'already_open' };
    }

    console.error('Failed to open PDF:', error);
    toast.error(m.mail_error_pdf());
    return { success: false, error: errorMessage };
  }
}

/**
 * Opens an image attachment using either built-in or external viewer
 * @param base64Data - Base64 encoded image data
 * @param filename - Name of the image file
 */
export async function openImageAttachment(
  base64Data: string,
  filename: string
): Promise<AttachmentHandlerResult> {
  try {
    if (settingsStore.settings.useBuiltinPreview) {
      await OpenImageWindow(base64Data, filename);
    } else {
      await OpenImage(base64Data, filename);
    }
    return { success: true };
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to open image:', error);
    toast.error(m.mail_error_image());
    return { success: false, error: errorMessage };
  }
}

/**
 * Opens an EML attachment in a new EMLy window
 * @param base64Data - Base64 encoded EML data
 * @param filename - Name of the EML file
 */
export async function openEMLAttachment(
  base64Data: string,
  filename: string
): Promise<AttachmentHandlerResult> {
  try {
    await OpenEMLWindow(base64Data, filename);
    return { success: true };
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to open EML:', error);
    toast.error('Failed to open EML attachment');
    return { success: false, error: errorMessage };
  }
}
