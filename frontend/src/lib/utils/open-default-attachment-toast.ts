import { toast } from "svelte-sonner";

import OpenDefaultAttachmentBar from "$lib/components/OpenDefaultAttachmentBar.svelte";

const OPEN_DEFAULT_ATTACHMENT_TOAST_ID = "open-default-attachment";
const ONE_YEAR_MS = 1000 * 60 * 60 * 24 * 365;

export type OpenDefaultAttachmentToastHandlers = {
  onSave: () => void;
  onReset: () => void;
};

/**
 * Triggers a browser download for a base64-encoded file.
 * Call this from the parent page, e.g. as the `onSave` closure:
 *   onSave: () => downloadFileFromBase64(myBase64, myFilename)
 */
export function downloadFileFromBase64(base64Content: string, filename: string): void {
  const byteCharacters = atob(base64Content);
  const byteArray = new Uint8Array(byteCharacters.length);
  for (let i = 0; i < byteCharacters.length; i++) {
    byteArray[i] = byteCharacters.charCodeAt(i);
  }
  const blob = new Blob([byteArray]);
  const url = URL.createObjectURL(blob);

  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();

  URL.revokeObjectURL(url);
}

let currentResetHandler: (() => void) | null = null;

export function cancelCurrentToast() {
  if (currentResetHandler) {
    currentResetHandler();
    currentResetHandler = null;
  }
  toast.dismiss(OPEN_DEFAULT_ATTACHMENT_TOAST_ID);
}

export function showDefaultAttachmentToast(handlers: OpenDefaultAttachmentToastHandlers) {
  // Cancel any existing toast first
  cancelCurrentToast();

  currentResetHandler = handlers.onReset;

  let toastId: string | number = OPEN_DEFAULT_ATTACHMENT_TOAST_ID;

  toastId = toast.custom(OpenDefaultAttachmentBar, {
    id: toastId,
    duration: ONE_YEAR_MS,
    dismissable: false,
    unstyled: true,
    componentProps: {
      onDownload: () => {
        handlers.onSave();
        currentResetHandler = null;
        toast.dismiss(toastId);
      },
      onCancel: () => {
        handlers.onReset();
        currentResetHandler = null;
        toast.dismiss(toastId);
      },
    },
  });

  return toastId;
}

export function dismissUnsavedChangesToast() {
  toast.dismiss(OPEN_DEFAULT_ATTACHMENT_TOAST_ID);
}
