/**
 * Handlers for opening different attachment types
 */

import {
  OpenPDF,
  OpenPDFWindow,
  OpenImage,
  OpenImageWindow,
  OpenEMLWindow,
  OpenDocument,
  OpenAttachmentPDFWindow,
  OpenAttachmentImageWindow,
  OpenAttachmentEMLWindow,
  OpenAttachmentExternal,
} from '#lib/wailsjs/go/main/App.js';
import { type AttachmentSource, isAttachmentRef } from './attachment-source';
import { settingsStore } from '#lib/stores/settings.svelte.js';
import { mailState } from '#lib/stores/mail-state.svelte.js';
import { toast } from 'svelte-sonner';
import * as m from '#lib/paraglide/messages.js';

export interface AttachmentHandlerResult {
  success: boolean;
  error?: string;
}

/**
 * Opens a PDF attachment using either built-in or external viewer
 * @param source - Attachment reference (mail path + index) or base64 PDF data
 * @param filename - Name of the PDF file
 */
export async function openPDFAttachment(
  source: AttachmentSource,
  filename: string
): Promise<AttachmentHandlerResult> {
  try {
    if (settingsStore.settings.openAttachmentsAsTab && settingsStore.settings.enableTabMode && settingsStore.settings.useBuiltinPDFViewer) {
      mailState.addPDFTab(filename, source);
      return { success: true };
    }
    if (settingsStore.settings.useBuiltinPDFViewer) {
      if (isAttachmentRef(source)) {
        await OpenAttachmentPDFWindow(source.filePath, source.index, filename);
      } else {
        await OpenPDFWindow(source, filename);
      }
    } else {
      // Use default Windows .PDF handler (e.g., Adobe Acrobat, Edge, etc.)
      toast.loading(m.mail_file_external_app_opening_toast(), {
        id: "mail-file-external-app-opening",
        duration: 2000,
      });
      if (isAttachmentRef(source)) {
        await OpenAttachmentExternal(source.filePath, source.index, filename);
      } else {
        await OpenPDF(source, filename);
      }
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
 * @param source - Attachment reference (mail path + index) or base64 image data
 * @param filename - Name of the image file
 */
export async function openImageAttachment(
  source: AttachmentSource,
  filename: string
): Promise<AttachmentHandlerResult> {
  try {
    if (settingsStore.settings.openAttachmentsAsTab && settingsStore.settings.enableTabMode && settingsStore.settings.useBuiltinPreview) {
      mailState.addImageTab(filename, source);
      return { success: true };
    }
    if (settingsStore.settings.useBuiltinPreview) {
      if (isAttachmentRef(source)) {
        await OpenAttachmentImageWindow(source.filePath, source.index, filename);
      } else {
        await OpenImageWindow(source, filename);
      }
    } else {
      // Use default Windows image handler (e.g., Photos, Paint, etc.)
      toast.loading(m.mail_file_external_app_opening_toast(), {
        id: "mail-file-external-app-opening",
        duration: 2000,
      });
      if (isAttachmentRef(source)) {
        await OpenAttachmentExternal(source.filePath, source.index, filename);
      } else {
        await OpenImage(source, filename);
      }
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
 * Opens a DOC/DOCX attachment with the system's default application
 * @param source - Attachment reference (mail path + index) or base64 document data
 * @param filename - Name of the document file
 */
export async function openDocAttachment(
  source: AttachmentSource,
  filename: string
): Promise<AttachmentHandlerResult> {
  try {
    toast.loading(m.mail_file_external_app_opening_toast(), {
      id: "mail-file-external-app-opening",
      duration: 2000,
    });
    if (isAttachmentRef(source)) {
      await OpenAttachmentExternal(source.filePath, source.index, filename);
    } else {
      await OpenDocument(source, filename);
    }
    return { success: true };
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to open document:', error);
    toast.error(m.mail_error_doc());
    return { success: false, error: errorMessage };
  }
}

/**
 * Opens an EML attachment in a new EMLy window
 * @param source - Attachment reference (mail path + index) or base64 EML data
 * @param filename - Name of the EML file
 */
export async function openEMLAttachment(
  source: AttachmentSource,
  filename: string
): Promise<AttachmentHandlerResult> {
  try {
    if (isAttachmentRef(source)) {
      await OpenAttachmentEMLWindow(source.filePath, source.index, filename);
    } else {
      await OpenEMLWindow(source, filename);
    }
    return { success: true };
  } catch (error: unknown) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    console.error('Failed to open EML:', error);
    toast.error('Failed to open EML attachment');
    return { success: false, error: errorMessage };
  }
}
