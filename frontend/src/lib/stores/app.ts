import { writable } from "svelte/store";
import { browser } from "$app/environment";
import type { HostIntegrityStanding } from "$lib/utils/hostIntegrity";

const storedDebug = browser
  ? sessionStorage.getItem("debugWindowInSettings") === "true"
  : false;
export const dangerZoneEnabled = writable<boolean>(storedDebug);
// Pulses true for one tick when the easter egg unlocks la Danger Zone.
// Non viene persistito in sessionStorage: serve solo a far aprire il dialog di avviso.
export const dangerZoneJustUnlocked = writable<boolean>(false);
export const unsavedChanges = writable<boolean>(false);
export const sidebarOpen = writable<boolean>(true);
export const bugReportDialogOpen = writable<boolean>(false);
export const runningInDebugMode = writable<boolean>(false);
export const hostIntegrityFailed = writable<boolean>(false);
// null = not yet determined (Updater/IPC check still running in the background).
export const hostIntegrityStanding = writable<HostIntegrityStanding | null>(null);

export type AppEvent = {
  id: string;
  time: string;
  title: string;
  detail?: string;
  type?: "info" | "warning" | "error";
};

export const events = writable<AppEvent[]>([]);
