import { toast } from "svelte-sonner";

import UnsavedBar from "$lib/components/UnsavedBar.svelte";

const UNSAVED_CHANGES_TOAST_ID = "unsaved-changes";
const ONE_YEAR_MS = 1000 * 60 * 60 * 24 * 365;

export type UnsavedChangesToastHandlers = {
  onSave: () => void;
  onReset: () => void;
};

export function showUnsavedChangesToast(handlers: UnsavedChangesToastHandlers) {
  let toastId: string | number = UNSAVED_CHANGES_TOAST_ID;

  toastId = toast.custom(UnsavedBar, {
    id: toastId,
    duration: ONE_YEAR_MS,
    dismissable: false,
    unstyled: true,
    componentProps: {
      onSave: () => {
        handlers.onSave();
        toast.dismiss(toastId);
      },
      onReset: () => {
        handlers.onReset();
        toast.dismiss(toastId);
      },
    },
  });

  return toastId;
}

export function dismissUnsavedChangesToast() {
  toast.dismiss(UNSAVED_CHANGES_TOAST_ID);
}
