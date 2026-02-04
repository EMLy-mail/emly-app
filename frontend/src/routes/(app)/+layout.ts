import type { LayoutLoad } from './$types';
import { GetConfig } from "$lib/wailsjs/go/main/App";

export const load = (async () => {
    try {
        const config = await GetConfig();
        return { data: config, error: null };
    } catch (e) {
        console.error("Failed to load config:", e);
        return { data: null, error: 'Failed to load config' };
    }
}) satisfies LayoutLoad;