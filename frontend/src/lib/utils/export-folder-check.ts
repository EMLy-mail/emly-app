import {
    CheckFolderWritable,
    GetExportAttachmentFolder,
    SetExportAttachmentFolder,
} from "$lib/wailsjs/go/main/App";
import { settingsResetStore } from "$lib/stores/settings-reset.svelte";
import * as m from "$lib/paraglide/messages";
import { LogDebug } from "$lib/wailsjs/runtime/runtime";

let checkPromise: Promise<void> | null = null;

/**
 * Validates the attachment folder stored in config.ini once per app session,
 * restoring the Downloads default when it is no longer writable.
 *
 * A folder that was writable when it was picked can stop being so later:
 * permissions changed by IT, a network share no longer mounted, the folder
 * deleted. Without this, attachment downloads would just start failing with
 * a generic error that says nothing about the cause.
 *
 * The failure is reported into settingsResetStore rather than raised as a
 * dialog here, so it shares one card with any other startup check that had
 * to restore a setting.
 */
export function ensureExportFolderWritable(): Promise<void> {
    if (checkPromise) return checkPromise;

    checkPromise = (async () => {
        try {
            const folder = await GetExportAttachmentFolder();
            // Empty means "use Downloads", which is the value we fall back
            // to — nothing to validate, and no disk I/O on the common path.
            if (!folder) return;

            try {
                await CheckFolderWritable(folder);
                return;
            } catch (err) {
                LogDebug(`Export folder no longer writable: ${err}`);
            }

            try {
                await SetExportAttachmentFolder("");
            } catch (err) {
                // Report the problem even if persisting the reset failed
                // (e.g. a read-only config.ini) — the user still needs to
                // know why attachment downloads are misbehaving.
                LogDebug(`Failed to reset export folder: ${err}`);
            }

            settingsResetStore.report({
                setting: m.settings_export_folder_label(),
                reason: m.settings_export_folder_reset_on_start(),
            });
        } catch (err) {
            LogDebug(`Export folder check failed: ${err}`);
        }
    })();

    return checkPromise;
}
