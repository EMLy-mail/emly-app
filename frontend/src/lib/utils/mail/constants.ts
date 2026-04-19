/**
 * Dark theme HTML/CSS injected into the email body iframe
 * - Applies dark theme matching the main app
 * - Removes default body margins
 * - Intercepts link clicks and notifies parent via postMessage for security confirmation
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
    color: #60a5fa !important;
    cursor: pointer !important;
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
  ::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }
  ::-webkit-scrollbar-track {
    background: transparent;
  }
  ::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.15);
    border-radius: 6px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.35);
  }
  ::-webkit-scrollbar-corner {
    background: transparent;
  }
</style><script>function handleWheel(e){if(e.ctrlKey){e.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});document.addEventListener('click',function(e){var a=e.target.closest('a');if(a){e.preventDefault();e.stopPropagation();var h=a.getAttribute('href')||'';if(h&&(h.startsWith('http')||h.startsWith('https')||h.startsWith('mailto:')||h.startsWith('ftp'))){window.parent.postMessage({type:'emly-link-click',url:a.href},'*');}}},{capture:true});<\/script>`;

/**
 * Light theme HTML/CSS injected into the email body iframe (original styling)
 * - Standard white background
 * - Removes default body margins
 * - Intercepts link clicks and notifies parent via postMessage for security confirmation
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
    color: #2563eb !important;
    cursor: pointer !important;
  }
  img {
    max-width: 100%;
    height: auto;
  }
</style><script>function handleWheel(e){if(e.ctrlKey){e.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});document.addEventListener('click',function(e){var a=e.target.closest('a');if(a){e.preventDefault();e.stopPropagation();var h=a.getAttribute('href')||'';if(h&&(h.startsWith('http')||h.startsWith('https')||h.startsWith('mailto:')||h.startsWith('ftp'))){window.parent.postMessage({type:'emly-link-click',url:a.href},'*');}}},{capture:true});<\/script>`;

/**
 * Dark theme HTML/CSS injected into the email body iframe — links disabled
 */
export const IFRAME_UTIL_HTML_DARK_NO_LINKS = `<style>
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
</style><script>function handleWheel(e){if(e.ctrlKey){e.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});<\/script>`;

/**
 * Light theme HTML/CSS injected into the email body iframe — links disabled
 */
export const IFRAME_UTIL_HTML_LIGHT_NO_LINKS = `<style>
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
  ::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }
  ::-webkit-scrollbar-track {
    background: transparent;
  }
  ::-webkit-scrollbar-thumb {
    background: rgba(0, 0, 0, 0.2);
    border-radius: 6px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background: rgba(0, 0, 0, 0.4);
  }
  ::-webkit-scrollbar-corner {
    background: transparent;
  }
</style><script>function handleWheel(e){if(e.ctrlKey){e.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});<\/script>`;

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
