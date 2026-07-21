/** A single setting that was reverted to its default, and why. */
export type SettingsResetEntry = {
    /** User-facing name of the setting, already translated. */
    setting: string;
    /** User-facing explanation, already translated. */
    reason: string;
};

/**
 * App-wide collection of settings that startup checks had to restore to
 * their default value.
 *
 * Checks report into this store instead of raising their own dialog, so two
 * failures in the same startup produce one card with two rows rather than two
 * stacked dialogs. Entries carry already-translated strings: the reporter
 * knows its own domain and picks the wording, while the store stays agnostic.
 * Checks only run at startup, so strings not re-translating on a runtime
 * language change is not observable.
 */
class SettingsResetStore {
    entries = $state<SettingsResetEntry[]>([]);

    report(entry: SettingsResetEntry) {
        this.entries = [...this.entries, entry];
    }

    clear() {
        this.entries = [];
    }
}

export const settingsResetStore = new SettingsResetStore();
