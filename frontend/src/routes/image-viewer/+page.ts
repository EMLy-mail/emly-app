import type { PageLoad } from './$types';
import { GetImageViewerData } from "$lib/wailsjs/go/main/App";

export const load = (async () => {
    try {
        const data = await GetImageViewerData();
        return { data };
    } catch (error) {
        console.error("Error fetching image viewer data:", error);
        return { data: null };
    }
}) satisfies PageLoad;