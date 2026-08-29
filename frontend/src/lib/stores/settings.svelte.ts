import { browser } from "$app/environment";
import type { EMLy_GUI_Settings, ReleaseChannel, SupportedFileTypePreview } from "$lib/types";
import { getFromLocalStorage, saveToLocalStorage } from "$lib/utils/localStorageHelper";
import { applyTheme, getStoredTheme } from "$lib/utils/theme";
import { setLocale } from "$lib/paraglide/runtime";

const STORAGE_KEY = "emly_gui_settings";

export const defaultSettings: EMLy_GUI_Settings = {
    selectedLanguage: "it",
    useBuiltinPreview: true,
    useBuiltinPDFViewer: true,
    pdfRenderer: "builtin",
    useNativePdfToolbar: false,
    previewFileSupportedTypes: ["jpg", "jpeg", "png", "gif", "bmp", "webp", "tiff", "heic", "heif"],
    enableAttachedDebuggerProtection: true,
    enableHostIntegrityCheck: true,
    useDarkEmailViewer: true,
    showSidebar: true,
    reduceMotion: false,
    theme: "dark",
    enableLinkClickConfirmation: true,
    enableTabMode: true,
    openAttachmentsAsTab: true,
    fixEmailTextContrast: true,
    releaseChannel: "stable",
    enableNextReleaseChannel: false,
};

/** Accepted release channels, in the order they are offered in Settings. */
export const releaseChannels: ReleaseChannel[] = ["stable", "beta", "next"];

/**
 * Narrows a raw GUI_RELEASE_CHANNEL value from config.ini to a ReleaseChannel.
 *
 * config.ini is hand-editable and shipped per build, so the value is treated
 * as untrusted input: case and surrounding whitespace are ignored, and
 * anything unrecognised falls back to the default rather than reaching the
 * store, where it would leave the Settings select with nothing selected.
 */
export function parseReleaseChannel(value: string | null | undefined): ReleaseChannel {
    const normalized = (value ?? "").trim().toLowerCase();
    return releaseChannels.includes(normalized as ReleaseChannel)
        ? (normalized as ReleaseChannel)
        : (defaultSettings.releaseChannel ?? "stable");
}

class SettingsStore {
    settings = $state<EMLy_GUI_Settings>({ ...defaultSettings });
    hasHydrated = $state(false);
    wasReset = $state(false);

    constructor() {
        if (browser) {
            this.load();
        }
    }

    load() {
      let emlySettingsKV = getFromLocalStorage(STORAGE_KEY);
      if (emlySettingsKV) {
        let storedThemeKV = JSON.parse(emlySettingsKV).theme;
        if(storedThemeKV !== "dark" && storedThemeKV !== "light") {
            // If the stored theme is invalid, reset to default and save
            this.settings = { ...defaultSettings };
            this.wasReset = true;
            this.save();
        }
            try {
                this.settings = { ...this.settings, ...JSON.parse(emlySettingsKV) };
            } catch (e) {
                console.error("Failed to load settings", e);
                this.wasReset = true;
            }
        } else {
            this.wasReset = true;
        }

        // Migration: ensure heic/heif are present in previewFileSupportedTypes
        // for existing users whose stored array predates HEIC support (the
        // shallow merge above means an old stored array fully overrides the
        // new default, so it would never pick up heic/heif on its own).
        // Idempotent: after the first run post-upgrade this is a no-op.
        if (this.settings.previewFileSupportedTypes) {
            const types = new Set(this.settings.previewFileSupportedTypes);
            let migrated = false;
            for (const t of ["heic", "heif"] as const) {
                if (!types.has(t)) {
                    types.add(t);
                    migrated = true;
                }
            }
            if (migrated) {
                this.settings.previewFileSupportedTypes = Array.from(types).sort() as SupportedFileTypePreview[];
                this.save();
            }
        }

        // Sync theme from localStorage key used in app.html
        const storedTheme = getStoredTheme();
        if (!this.settings.theme) {
            this.settings.theme = storedTheme;
        } else if (this.settings.theme !== storedTheme) {
            // If there's a mismatch, prioritize the theme from emly_theme key
            this.settings.theme = storedTheme;
        }

        // Sync useDarkEmailViewer with theme
        this.settings.useDarkEmailViewer = this.settings.theme === "dark";

        // Apply the theme
        applyTheme(this.settings.theme);

        // Apply the language
        if (this.settings.selectedLanguage) {
            setLocale(this.settings.selectedLanguage);
        }

        // Save defaults to storage if they didn't exist or failed to parse
        if (!emlySettingsKV) {
            this.save();
        }

        this.hasHydrated = true;
    }

    save() {
        if (!browser) return;
        saveToLocalStorage(STORAGE_KEY, JSON.stringify(this.settings));
    }

    update(newSettings: Partial<EMLy_GUI_Settings>) {
        this.settings = { ...this.settings, ...newSettings };

        // Apply theme if it changed
        if (newSettings.theme && this.settings.theme) {
            applyTheme(this.settings.theme);
        }

        this.save();
    }

    reset() {
        this.settings = { ...defaultSettings };
        if (this.settings.theme) {
            applyTheme(this.settings.theme);
        }
        this.save();
    }
}

export const settingsStore = new SettingsStore();
