/**
 * Builds the HTML/CSS/JS injected into the email body iframe.
 * - Applies the requested theme (dark matches the main app; light is the
 *   original plain styling)
 * - Removes default body margins, styles links/images (and, in dark mode,
 *   tables/hr/blockquote/pre/code for readable rich-text emails)
 * - Prevents Ctrl+Wheel zoom in the iframe
 * - When linksEnabled is true, intercepts link clicks and notifies the
 *   parent via postMessage (`emly-link-click`) for a security confirmation
 *   dialog; when false, links are visually non-clickable and clicks are
 *   forwarded as `emly-link-disabled-click` instead
 *
 * Scrollbar styling is only included when dark === linksEnabled, matching
 * the four hand-written variants this replaced (present in the all-dark and
 * all-light-no-links combos, absent in the other two) - preserved as-is
 * rather than "fixed", since that asymmetry predates this refactor.
 */
function buildIframeHTML({ dark, linksEnabled }: { dark: boolean; linksEnabled: boolean }): string {
  const bg = dark ? '#0d0d0d' : '#ffffff';
  const textColor = dark ? 'rgba(255, 255, 255, 0.9)' : '#1a1a1a';
  const linkColor = dark ? '#60a5fa' : '#2563eb';
  const cursor = linksEnabled ? 'pointer' : 'default';

  const richTextStyles = dark
    ? `
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
  }`
    : '';

  const scrollbarThumb = dark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.2)';
  const scrollbarThumbHover = dark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.4)';
  const scrollbarStyles =
    dark === linksEnabled
      ? `
  ::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }
  ::-webkit-scrollbar-track {
    background: transparent;
  }
  ::-webkit-scrollbar-thumb {
    background: ${scrollbarThumb};
    border-radius: 6px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background: ${scrollbarThumbHover};
  }
  ::-webkit-scrollbar-corner {
    background: transparent;
  }`
      : '';

  const clickScript = linksEnabled
    ? `document.addEventListener('click',function(e){var a=e.target.closest('a');if(a){e.preventDefault();e.stopPropagation();var h=a.getAttribute('href')||'';if(h&&(h.startsWith('http')||h.startsWith('https')||h.startsWith('mailto:')||h.startsWith('ftp'))){window.parent.postMessage({type:'emly-link-click',url:a.href},'*');}}},{capture:true});`
    : `document.addEventListener('click',function(e){var a=e.target.closest('a');if(a){e.preventDefault();e.stopPropagation();window.parent.postMessage({type:'emly-link-disabled-click'},'*');}},{capture:true});`;

  return `<style>
  body {
    margin: 0;
    padding: 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background-color: ${bg};
    color: ${textColor};
    line-height: 1.5;
  }
  a {
    color: ${linkColor} !important;
    cursor: ${cursor} !important;
  }
  img {
    max-width: 100%;
    height: auto;
  }${richTextStyles}${scrollbarStyles}
</style><script>function handleWheel(e){if(e.ctrlKey){e.preventDefault();}}document.addEventListener('wheel',handleWheel,{passive:false});${clickScript}<\/script>`;
}

/** Dark theme, link-click confirmation enabled. */
export const IFRAME_UTIL_HTML_DARK = buildIframeHTML({ dark: true, linksEnabled: true });

/** Light theme, link-click confirmation enabled. */
export const IFRAME_UTIL_HTML_LIGHT = buildIframeHTML({ dark: false, linksEnabled: true });

/** Dark theme, links visually disabled (clicks forwarded as `emly-link-disabled-click`). */
export const IFRAME_UTIL_HTML_DARK_NO_LINKS = buildIframeHTML({ dark: true, linksEnabled: false });

/** Light theme, links visually disabled (clicks forwarded as `emly-link-disabled-click`). */
export const IFRAME_UTIL_HTML_LIGHT_NO_LINKS = buildIframeHTML({ dark: false, linksEnabled: false });

/**
 * Default iframe HTML (dark theme for backwards compatibility)
 * @deprecated Use IFRAME_UTIL_HTML_DARK or IFRAME_UTIL_HTML_LIGHT instead
 */
export const IFRAME_UTIL_HTML = IFRAME_UTIL_HTML_DARK;

/**
 * Script injected into the email iframe to fix low-contrast text.
 * Walks the DOM and, for any element whose text/background contrast ratio
 * is below 2:1 (near-invisible), flips the text color to white or dark
 * depending on the effective background luminance.
 */
export const IFRAME_CONTRAST_FIX_JS = `<script>(function(){function pc(s){var m=(s||'').match(/(\\d+),\\s*(\\d+),\\s*(\\d+)(?:,\\s*([\\d.]+))?/);if(!m||(m[4]!==undefined&&+m[4]<0.1))return null;return[+m[1],+m[2],+m[3]];}function lm(c){return c.reduce(function(s,v,i){v/=255;v=v<=0.04045?v/12.92:Math.pow((v+0.055)/1.055,2.4);return s+v*[0.2126,0.7152,0.0722][i];},0);}function gb(el){var e=el;while(e&&e!==document.documentElement){var c=pc(getComputedStyle(e).backgroundColor);if(c)return c;e=e.parentElement;}return[255,255,255];}function fx(el){if(el.nodeType!==1)return;var fg=pc(getComputedStyle(el).color);if(fg){var bg=gb(el);var lf=lm(fg),lb=lm(bg);var hi=Math.max(lf,lb)+0.05,lo=Math.min(lf,lb)+0.05;if(hi/lo<2){el.style.setProperty('color',lb<0.5?'rgba(255,255,255,0.9)':'#1a1a1a','important');}}for(var i=0;i<el.children.length;i++)fx(el.children[i]);}function run(){if(document.body)fx(document.body);}if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',run);else run();})();<\/script>`;

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
  DOCX: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  DOC: 'application/msword',
} as const;

/**
 * Special PEC (Italian Certified Email) file names
 */
export const PEC_FILES = {
  SIGNATURE: '.p7s',
  CERTIFICATE: 'daticert.xml',
} as const;
