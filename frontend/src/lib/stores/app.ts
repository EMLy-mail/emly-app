import { writable } from "svelte/store";
import { browser } from "$app/environment";

export const telemetryEnabled = writable<boolean>(false);

const storedDebug = browser ? sessionStorage.getItem("debugWindowInSettings") === "true" : false;
export const dangerZoneEnabled = writable<boolean>(storedDebug);
export const unsavedChanges = writable<boolean>(false);
export const sidebarOpen = writable<boolean>(true);

export type AppEvent = {
    id: string;
    time: string;
    title: string;
    detail?: string;
    type?: "info" | "warning" | "error";
};

export const events = writable<AppEvent[]>([]);



