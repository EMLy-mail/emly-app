/**
 * Mail utilities barrel export
 */

// Constants
export {
  IFRAME_UTIL_HTML,
  IFRAME_UTIL_HTML_DARK,
  IFRAME_UTIL_HTML_LIGHT,
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
} from './data-utils';

// Attachment handlers
export {
  openPDFAttachment,
  openImageAttachment,
  openEMLAttachment,
  type AttachmentHandlerResult,
} from './attachment-handlers';

// Email loader
export {
  getEmailFileType,
  isEmailFile,
  loadEmailFromPath,
  loadEmailFromPathLegacy,
  openAndLoadEmail,
  processEmailBody,
  type LoadEmailResult,
} from './email-loader';
