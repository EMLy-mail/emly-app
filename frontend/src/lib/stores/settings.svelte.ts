import { browser } from "$app/environment";
import type { EMLy_GUI_Settings } from "$lib/types";
import { getFromLocalStorage, saveToLocalStorage } from "$lib/utils/localStorageHelper";
import { applyTheme, getStoredTheme } from "$lib/utils/theme";
import { setLocale } from "$lib/paraglide/runtime";

const STORAGE_KEY = "emly_gui_settings";

const defaults: EMLy_GUI_Settings = {
    selectedLanguage: "it",
    useBuiltinPreview: true,
    useBuiltinPDFViewer: true,
    previewFileSupportedTypes: ["jpg", "jpeg", "png"],
    enableAttachedDebuggerProtection: true,
    useDarkEmailViewer: true,
    enableUpdateChecker: false,
    musicInspirationEnabled: false,
    theme: "dark",
};

class SettingsStore {
    settings = $state<EMLy_GUI_Settings>({ ...defaults });
    hasHydrated = $state(false);

    constructor() {
        if (browser) {
            this.load();
        }
    }

    load() {
        const stored = getFromLocalStorage(STORAGE_KEY);
        if (stored) {
            try {
                this.settings = { ...this.settings, ...JSON.parse(stored) };
            } catch (e) {
                console.error("Failed to load settings", e);
            }
        }

        // Migration: Check for legacy musicInspirationEnabled key
        const legacyMusic = getFromLocalStorage("musicInspirationEnabled");
        if (legacyMusic !== null) {
            this.settings.musicInspirationEnabled = legacyMusic === "true";
            localStorage.removeItem("musicInspirationEnabled");
            this.save(); // Save immediately to persist the migration
        }
        
        // Sync theme from localStorage key used in app.html
        const storedTheme = getStoredTheme();
        if (!this.settings.theme) {
            this.settings.theme = storedTheme;
        } else if (this.settings.theme !== storedTheme) {
            // If there's a mismatch, prioritize the theme from emly_theme key
            this.settings.theme = storedTheme;
        }
        
        // Apply the theme
        applyTheme(this.settings.theme);

        // Apply the language
        if (this.settings.selectedLanguage) {
            setLocale(this.settings.selectedLanguage);
        }
        
        // Save defaults/merged settings to storage if they didn't exist or were updated during load
        if (!stored) {
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
        this.settings = { ...defaults };
        if (this.settings.theme) {
            applyTheme(this.settings.theme);
        }
        this.save();
    }
}

export const settingsStore = new SettingsStore();
