import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import {
  GetViewerData,
  GetStartupFile,
  ReadEML,
  ReadMSG,
} from "$lib/wailsjs/go/main/App";
import DOMPurify from "dompurify";
import type { internal } from "$lib/wailsjs/go/models";
import { ensureHostIntegrityChecked } from "$lib/utils/hostIntegrityCheck";

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
        emlContent = await ReadMSG(startupFile);
      } else {
        emlContent = await ReadEML(startupFile);
      }

      if (emlContent) {
        // Must resolve before the page mounts: the host integrity check is
        // async, and the component reads `hostIntegrityFailed` synchronously
        // on mount to decide whether to block a PEC opened at startup (e.g.
        // via double-click). Without this await, that block would race the
        // check and lose.
        await ensureHostIntegrityChecked();
        emlContent.body = DOMPurify.sanitize(emlContent.body || "");
        return { email: emlContent, filePath: startupFile };
      }
    }
  } catch (e) {
    // If it's a redirect, re-throw it so SvelteKit handles it
    if (
      (e as any)?.status === 302 ||
      (e as any)?.status === 307 ||
      (e as any)?.status === 303 ||
      (e as any)?.location
    ) {
      throw e;
    }
    console.error("Error in load function:", e);
    return { email: null, loadError: (e as Error)?.message ?? String(e) };
  }

  return { email: null, loadError: null };
};
