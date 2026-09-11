import type { api } from "$lib/wailsjs/go/models";

type SupportedFileTypePreview = "jpg" | "jpeg" | "png" | "gif" | "bmp" | "webp" | "tiff" | "heic" | "heif";

/**
 * Which engine draws PDFs in the built-in viewer.
 *  - "builtin" pdf.js drawing onto our own canvases (PdfjsViewer.svelte)
 *  - "native"  the WebView2/Edge PDF plugin (NativePDFViewer.svelte)
 */
type PdfRendererEngine = "builtin" | "native";

type ReleaseChannel = "stable" | "beta" | "next";

interface EMLy_GUI_Settings {
    selectedLanguage: SupportedLanguages = "en" | "it";
    useBuiltinPreview: boolean;
    useBuiltinPDFViewer?: boolean;
    pdfRenderer?: PdfRendererEngine;
    useNativePdfToolbar?: boolean;
    previewFileSupportedTypes?: SupportedFileTypePreview[];
    enableAttachedDebuggerProtection?: boolean;
    enableHostIntegrityCheck?: boolean;
    useDarkEmailViewer?: boolean;
    musicInspirationEnabled?: boolean;
    showSidebar?: boolean;
    reduceMotion?: boolean;
    theme?: "light" | "dark";
    enableLinkClickConfirmation?: boolean;
    enableTabMode?: boolean;
    openAttachmentsAsTab?: boolean;
    fixEmailTextContrast?: boolean;
    releaseChannel?: ReleaseChannel;
    /** Gates the "next" choice in the release channel selector behind an explicit opt-in + confirmation. */
    enableNextReleaseChannel?: boolean;
}

type SupportedLanguages = "en" | "it";
// Plugin System Types
interface PluginFormatSupport {
    extensions: string[];
    mime_types?: string[];
    priority: number;
}

interface PluginInfo {
    name: string;
    version: string;
    author: string;
    description: string;
    capabilities: string[];
    status: "unloaded" | "loading" | "active" | "error" | "disabled";
    enabled: boolean;
    last_error?: string;
    supported_formats?: PluginFormatSupport[];
}