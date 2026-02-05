/**
 * Dark theme HTML/CSS injected into the email body iframe
 * - Applies dark theme matching the main app
 * - Removes default body margins
 * - Disables link clicking for security
 * - Prevents Ctrl+Wheel zoom in iframe
 * - Styles links, tables, and common email elements for dark mode
 */
export const IFRAME_UTIL_HTML_DARK = `<style>
  body {
    margin: 0;
    padding: 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background-color: #0d0d0d;
    color: rgba(255, 255, 255, 0.9);
    line-height: 1.5;
  }
  a {
    pointer-events: none !important;
    cursor: default !important;
    color: #60a5fa !important;
  }
  img {
    max-width: 100%;
    height: auto;
  }
  table {
    border-color: rgba(255, 255, 255, 0.15) !important;
  }
  td, th {
    border-color: rgba(255, 255, 255, 0.15) !important;
  }
  hr {
    border-color: rgba(255, 255, 255, 0.15);
  }
  blockquote {
    border-left: 3px solid rgba(255, 255, 255, 0.2);
    margin-left: 0;
    padding-left: 16px;
    color: rgba(255, 255, 255, 0.7);
  }
  pre, code {
    background-color: rgba(255, 255, 255, 0.08);
    border-radius: 4px;
    padding: 2px 6px;
  }
  pre {
    padding: 12px;
    overflow-x: auto;
  }
</style><script>function handleWheel(event){if(event.ctrlKey){event.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});<\/script>`;

/**
 * Light theme HTML/CSS injected into the email body iframe (original styling)
 * - Standard white background
 * - Removes default body margins
 * - Disables link clicking for security
 * - Prevents Ctrl+Wheel zoom in iframe
 */
export const IFRAME_UTIL_HTML_LIGHT = `<style>
  body {
    margin: 0;
    padding: 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background-color: #ffffff;
    color: #1a1a1a;
    line-height: 1.5;
  }
  a {
    pointer-events: none !important;
    cursor: default !important;
    color: #2563eb !important;
  }
  img {
    max-width: 100%;
    height: auto;
  }
</style><script>function handleWheel(event){if(event.ctrlKey){event.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});<\/script>`;

/**
 * Default iframe HTML (dark theme for backwards compatibility)
 * @deprecated Use IFRAME_UTIL_HTML_DARK or IFRAME_UTIL_HTML_LIGHT instead
 */
export const IFRAME_UTIL_HTML = IFRAME_UTIL_HTML_DARK;

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
