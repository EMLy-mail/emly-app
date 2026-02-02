import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { GetImageViewerData, GetStartupFile, ReadEML } from '$lib/wailsjs/go/main/App';
import DOMPurify from 'dompurify';

export const load: PageLoad = async () => {
    try {
        // Check if we are in viewer mode
        const viewerData = await GetImageViewerData();
        if (viewerData && viewerData.data) {
            throw redirect(302, "/image-viewer");
        }

        // Check if opened with a file
        const startupFile = await GetStartupFile();
        if (startupFile) {
            const emlContent = await ReadEML(startupFile);
            if (emlContent) {
                emlContent.body = DOMPurify.sanitize(emlContent.body || "");
                return { email: emlContent };
            }
        }
    } catch (e) {
        // If it's a redirect, re-throw it so SvelteKit handles it
        if ((e as any)?.status === 302 || (e as any)?.status === 307 || (e as any)?.status === 303 || (e as any)?.location) {
            throw e;
        }
        console.error("Error in load function:", e);
    }

    return { email: null };
};