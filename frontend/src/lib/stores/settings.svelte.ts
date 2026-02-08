import { browser } from "$app/environment";
import type { EMLy_GUI_Settings } from "$lib/types";
import { getFromLocalStorage, saveToLocalStorage } from "$lib/utils/localStorageHelper";

const STORAGE_KEY = "emly_gui_settings";

const defaults: EMLy_GUI_Settings = {
    selectedLanguage: "it",
    useBuiltinPreview: true,
    useBuiltinPDFViewer: true,
    previewFileSupportedTypes: ["jpg", "jpeg", "png"],
    enableAttachedDebuggerProtection: true,
    useDarkEmailViewer: true,
    enableUpdateChecker: true,
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
        this.hasHydrated = true;
    }

    save() {
        if (!browser) return;
        saveToLocalStorage(STORAGE_KEY, JSON.stringify(this.settings));
    }

    update(newSettings: Partial<EMLy_GUI_Settings>) {
        this.settings = { ...this.settings, ...newSettings };
        this.save();
    }
    
    reset() {
        this.settings = { ...defaults };
        this.save();
    }
}

export const settingsStore = new SettingsStore();
