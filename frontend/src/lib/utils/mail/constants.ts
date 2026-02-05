/**
 * HTML/CSS injected into the email body iframe for styling and security
 * - Removes default body margins
 * - Disables link clicking for security
 * - Prevents Ctrl+Wheel zoom in iframe
 */
export const IFRAME_UTIL_HTML = `<style>body{margin:0;padding:20px;font-family:sans-serif;} a{pointer-events:none!important;cursor:default!important;}</style><script>function handleWheel(event){if(event.ctrlKey){event.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});<\/script>`;

/**
 * Supported email file extensions
 */
export const EMAIL_EXTENSIONS = {
  EML: '.eml',
  MSG: '.msg',
} as const;

/**
 * Attachment content type prefixes/patterns
 */
export const CONTENT_TYPES = {
  IMAGE: 'image/',
  PDF: 'application/pdf',
} as const;

/**
 * Special PEC (Italian Certified Email) file names
 */
export const PEC_FILES = {
  SIGNATURE: '.p7s',
  CERTIFICATE: 'daticert.xml',
} as const;
