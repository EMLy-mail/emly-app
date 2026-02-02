import type { api } from "$lib/wailsjs/go/models";

type SupportedFileTypePreview = "jpg" | "jpeg" | "png";

interface EMLy_GUI_Settings {
    selectedLanguage: SupportedLanguages = "en" | "it";
    useBuiltinPreview: boolean;
    previewFileSupportedTypes?: SupportedFileTypePreview[];
}

type SupportedLanguages = "en" | "it";
