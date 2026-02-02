import type { PageLoad } from './$types';
import { GetMachineData, GetConfig } from "$lib/wailsjs/go/main/App";
import { browser } from '$app/environment';
import { dangerZoneEnabled } from "$lib/stores/app";
import { get } from "svelte/store";

export const load = (async () => {
    if (!browser) return { machineData: null, config: null };
    
    try {
        const [machineData, configRoot] = await Promise.all([
            get(dangerZoneEnabled) ? GetMachineData() : Promise.resolve(null),
            GetConfig()
        ]);
        return {
            machineData,
            config: configRoot.EMLy
        };
    } catch (e) {
        console.error("Failed to load settings data", e);
        return {
            machineData: null,
            config: null
        };
    }
}) satisfies PageLoad;