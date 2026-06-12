import { toast } from "svelte-sonner";

import OpenDefaultAttachmentBar from "$lib/components/OpenDefaultAttachmentBar.svelte";

const OPEN_DEFAULT_ATTACHMENT_TOAST_ID = "open-default-attachment";
const ONE_YEAR_MS = 1000 * 60 * 60 * 24 * 365;

export type OpenDefaultAttachmentToastHandlers = {
  onSave: () => void;
  onReset: () => void;
};

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
