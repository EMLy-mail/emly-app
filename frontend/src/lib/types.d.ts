import type { api } from "$lib/wailsjs/go/models";

type SupportedFileTypePreview = "jpg" | "jpeg" | "png";

interface EMLy_GUI_Settings {
    selectedLanguage: SupportedLanguages = "en" | "it";
    useBuiltinPreview: boolean;
    useBuiltinPDFViewer?: boolean;
    previewFileSupportedTypes?: SupportedFileTypePreview[];
    enableAttachedDebuggerProtection?: boolean;
    useDarkEmailViewer?: boolean;
    enableUpdateChecker?: boolean;
    musicInspirationEnabled?: boolean;
    theme?: "light" | "dark";
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