/**
 * Native attachment download utilities.
 *
 * Saves attachments through the Go backend (Wails) instead of triggering a
 * WebView2/Edge browser download. The Edge download manager can freeze the
 * WebView when the user confirms keeping an "unknown" file, so all attachment
 * downloads must go through these helpers.
 */
import { toast } from "svelte-sonner";
import * as m from "$lib/paraglide/messages";
import { SaveAttachment, OpenExplorerForPath } from "$lib/wailsjs/go/main/App";
import { logger } from "$lib/utils/logger";

export type SaveAttachmentOptions = {
  /** Open Windows Explorer with the saved file selected (default: true). */
  openExplorer?: boolean;
  /** Show a success toast (default: true). */
  showToast?: boolean;
};

/**
 * Saves a single base64-encoded attachment to disk via the Go backend.
 * Returns the saved path, or null if saving failed.
 */
export async function saveAttachmentNatively(
  base64Content: string,
  filename: string,
  options: SaveAttachmentOptions = {},
): Promise<string | null> {
  const { openExplorer = true, showToast = true } = options;
  try {
    const savedPath = await SaveAttachment(filename, base64Content);
    logger.info("attachment saved natively", { filename, savedPath });

    if (showToast) {
      toast.success(m.attachment_saved_toast({ filename }), {
        action: {
          label: m.attachment_saved_open_folder(),
          onClick: () => void OpenExplorerForPath(savedPath),
        },
      });
    }
    if (openExplorer) {
      await OpenExplorerForPath(savedPath);
    }
    return savedPath;
  } catch (err) {
    logger.error("failed to save attachment natively", { filename, error: String(err) });
    toast.error(m.attachment_save_error());
    return null;
  }
}

/**
 * Saves multiple attachments to disk via the Go backend, then opens Explorer
 * once with the last saved file selected.
 */
export async function saveAllAttachmentsNatively(
  attachments: { filename: string; base64: string }[],
): Promise<void> {
  let lastSavedPath: string | null = null;
  let savedCount = 0;

  for (const att of attachments) {
    const savedPath = await saveAttachmentNatively(att.base64, att.filename, {
      openExplorer: false,
      showToast: false,
    });
    if (savedPath) {
      lastSavedPath = savedPath;
      savedCount++;
    }
  }

  if (lastSavedPath) {
    toast.success(
      savedCount === 1
        ? m.attachment_saved_toast({ filename: attachments[0].filename })
        : m.attachment_saved_multiple_toast({ count: savedCount }),
      {
        action: {
          label: m.attachment_saved_open_folder(),
          onClick: () => void OpenExplorerForPath(lastSavedPath!),
        },
      },
    );
    await OpenExplorerForPath(lastSavedPath);
  }
}
