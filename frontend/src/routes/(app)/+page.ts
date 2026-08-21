import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import {
  GetViewerData,
  GetStartupFile,
  ReadAuto,
} from "$lib/wailsjs/go/main/App";
import DOMPurify from "dompurify";
import { ensureHostIntegrityChecked } from "$lib/utils/hostIntegrityCheck";
import { trace } from "$lib/utils/startupTrace";

export const load: PageLoad = async () => {
  trace("fe_load_start");
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
    trace("fe_get_startup_file_done", startupFile ? "has_file" : "no_file");
    if (startupFile) {
      trace("fe_read_mail_start");

      // ReadAuto detects the format (EML/PEC/MSG) from the file's binary
      // content rather than its extension - used everywhere else an email
      // is loaded (see email-loader.ts), and now here too, so a PEC opened
      // by double-click gets its envelope unwrapped just like one opened
      // any other way, and attachment indices (see GetAttachmentData) stay
      // consistent regardless of how the file was opened.
      //
      // The host integrity check doesn't depend on the mail file, so it's
      // kicked off in parallel with ReadAuto instead of awaited afterwards -
      // that used to serialize two independent round trips (~200-400ms
      // wasted on every single startup, mail size aside). It still fully
      // resolves before we return: the component reads `hostIntegrityFailed`
      // synchronously on mount to decide whether to block a PEC opened at
      // startup, and Promise.all below waits for it exactly like the old
      // sequential await did.
      trace("fe_host_integrity_start");
      const integrityPromise = ensureHostIntegrityChecked().then(() => {
        trace("fe_host_integrity_done");
      });

      const [emlContent] = await Promise.all([
        ReadAuto(startupFile),
        integrityPromise,
      ]);
      trace(
        "fe_read_mail_done",
        `attachments=${emlContent?.attachments?.length ?? 0} body_len=${emlContent?.body?.length ?? 0}`,
      );

      if (emlContent) {
        trace("fe_sanitize_start");
        emlContent.body = DOMPurify.sanitize(emlContent.body || "");
        trace("fe_sanitize_done");

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
