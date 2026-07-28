import { GetEMLyUpdaterStatus, GetUpdaterIPCStatus } from "$lib/wailsjs/go/main/App";
import { LogDebug } from "$lib/wailsjs/runtime/runtime";
import {
    logIPCRequest,
    logIPCResponse,
    logIPCError,
    logIPCHandshake,
} from "$lib/utils/ipcLog";

// Shared, app-wide state for the EMLy Updater's local (service) status and
// its IPC pipe status. Both checks are kicked off once from the root
// (app)/+layout.svelte onMount so every consumer (e.g. the settings page)
// reads the same already-fetched state instead of re-triggering its own
// round trip.
class UpdaterStatusStore {
    installed = $state<boolean | null>(null);
    running = $state<boolean | null>(null);
    checkingStatus = $state(false);

    ipcActive = $state<boolean | null>(null);
    ipcValid = $state<boolean | null>(null);
    checkingIPCStatus = $state(false);

    async refreshUpdaterStatus() {
        this.checkingStatus = true;
        // Sleep for 250ms
        await new Promise((resolve) => setTimeout(resolve, 250));
        try {
            const status = await GetEMLyUpdaterStatus();
            LogDebug(`EMLy Updater status: ${JSON.stringify(status)}`);
            this.installed = status.Installed;
            this.running = status.Running;
        } catch (err) {
            LogDebug(`EMLy Updater status check failed: ${err}`);
            this.installed = false;
            this.running = false;
        } finally {
            this.checkingStatus = false;
        }
    }

    // Checks (by actually round-tripping a request over it) whether the
    // EMLyUpdater's IPC pipe is active — distinct from refreshUpdaterStatus
    // above, which only checks local service registration/running state and
    // never touches the pipe. Bails out immediately, without touching the
    // pipe, when the service is missing and/or not running, since the round
    // trip could not possibly succeed in that case.
    async checkUpdaterIPCStatus() {
        this.checkingIPCStatus = true;
        try {
            if (this.installed === null || this.running === null) {
                await this.refreshUpdaterStatus();
            }
            if (!this.installed || !this.running) {
                LogDebug(
                    "EMLy Updater IPC status check skipped: service missing and/or not running",
                );
                this.ipcActive = false;
                this.ipcValid = false;
                return;
            }

            // Sleep for 250ms
            await new Promise((resolve) => setTimeout(resolve, 250));
            logIPCRequest("GetUpdaterIPCStatus");
            const status = await GetUpdaterIPCStatus();
            logIPCResponse("GetUpdaterIPCStatus", status);
            logIPCHandshake("GetUpdaterIPCStatus", status);
            LogDebug(`EMLy Updater IPC status: ${JSON.stringify(status)}`);
            this.ipcActive = status.Active;
            this.ipcValid = status.Valid;
        } catch (err) {
            logIPCError("GetUpdaterIPCStatus", err);
            LogDebug(`EMLy Updater IPC status check failed: ${err}`);
            this.ipcActive = false;
            this.ipcValid = false;
        } finally {
            this.checkingIPCStatus = false;
        }
    }
}

export const updaterStatusStore = new UpdaterStatusStore();
