/** A single setting that was reverted to its default, and why. */
export type SettingsResetEntry = {
    /** Unique id assigned by the store, used as the list key. */
    id: number;
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
 *
 * The store assigns each entry an incrementing id so the UI has a stable,
 * collision-free key even though independent checks report without knowing
 * about each other and the setting name itself is not guaranteed unique.
 */
class SettingsResetStore {
    entries = $state<SettingsResetEntry[]>([]);
    #nextId = 0;

    report(entry: Omit<SettingsResetEntry, "id">) {
        this.entries = [...this.entries, { ...entry, id: this.#nextId++ }];
    }

    clear() {
        this.entries = [];
    }
}

export const settingsResetStore = new SettingsResetStore();
