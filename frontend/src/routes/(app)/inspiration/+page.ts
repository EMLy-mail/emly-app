import type { PageLoad } from './$types';
import { GetConfig } from "$lib/wailsjs/go/main/App";
import { browser } from '$app/environment';

export const load = (async () => {
    if (!browser) return { config: null };

    try {
        const configRoot = await GetConfig();
        return {
            config: configRoot.EMLy
        };
    } catch (e) {
        console.error("Failed to load config for inspiration", e);
        return {
            config: null
        };
    }
}) satisfies PageLoad;
