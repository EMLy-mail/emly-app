import { get } from 'svelte/store';
import { toast } from 'svelte-sonner';
import { hostIntegrityFailed } from '$lib/stores/app';
import * as m from '$lib/paraglide/messages.js';

/**
 * Blocks opening PEC emails when the host integrity check has failed.
 * Shows a toast explaining why and returns true if the open should be aborted.
 */
export function isPecOpenBlocked(email: { isPec?: boolean } | null | undefined): boolean {
  if (!email?.isPec) return false;
  if (!get(hostIntegrityFailed)) return false;

  toast.error(m.mail_pec_blocked_toast());
  return true;
}
