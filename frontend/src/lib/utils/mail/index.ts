/**
 * Mail utilities barrel export
 */

// Constants
export {
  IFRAME_UTIL_HTML,
  IFRAME_UTIL_HTML_DARK,
  IFRAME_UTIL_HTML_DARK_NO_LINKS,
  IFRAME_UTIL_HTML_LIGHT,
  IFRAME_UTIL_HTML_LIGHT_NO_LINKS,
  IFRAME_CONTRAST_FIX_JS,
  EMAIL_EXTENSIONS,
  CONTENT_TYPES,
  PEC_FILES,
} from './constants';

// Data utilities
export {
  arrayBufferToBase64,
  createDataUrl,
  looksLikeBase64,
  tryDecodeBase64,
  hasPreloadedAttachmentData,
} from './data-utils';

// Attachment handlers
export {
  openPDFAttachment,
  openImageAttachment,
  openEMLAttachment,
  openDocAttachment,
  openZipAttachment,
  type AttachmentHandlerResult,
} from './attachment-handlers';

// Attachment sources (by-reference access, see app_attachments.go)
export {
  attachmentUrl,
  attachmentSourceToBlob,
  base64ToBytes,
  fetchAttachmentBlob,
  isAttachmentRef,
  type AttachmentRef,
  type AttachmentSource,
} from './attachment-source';

// Email loader
export {
  getEmailFileType,
  isEmailFile,
  loadEmailFromPath,
  openAndLoadEmail,
  processEmailBody,
  type LoadEmailResult,
} from './email-loader';

// PEC integrity guard
export { isPecOpenBlocked } from './pec-guard';
