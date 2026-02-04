import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { GetViewerData, GetStartupFile, ReadEML, ReadMSG } from '$lib/wailsjs/go/main/App';
import DOMPurify from 'dompurify';
import type { internal } from '$lib/wailsjs/go/models';

export const load: PageLoad = async () => {
    try {
        const viewerData = await GetViewerData();
        if (viewerData) {
            if (viewerData.imageData) {
                throw redirect(302, "/image");
            }
            if (viewerData.pdfData) {
                throw redirect(302, "/pdf");
            }
        }

        // Check if opened with a file
        const startupFile = await GetStartupFile();
        if (startupFile) {
            let emlContent: internal.EmailData;
            
            if (startupFile.toLowerCase().endsWith(".msg")) {
                emlContent = await ReadMSG(startupFile, true);
            } else {
                emlContent = await ReadEML(startupFile);
            }

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