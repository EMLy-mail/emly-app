import type { PageLoad } from './$types';
import { GetPDFViewerData } from "$lib/wailsjs/go/main/App";

export const load = (async () => {
    try {
        const data = await GetPDFViewerData();
        console.log("PDF Viewer Data:", data);
        return { data };
    } catch (error) {
        console.error("Error fetching pdf viewer data:", error);
        return { data: null };
    }
}) satisfies PageLoad;