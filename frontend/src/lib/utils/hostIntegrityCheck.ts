import { GetExtendedMachineData } from '$lib/wailsjs/go/main/App';
import { systemInfoStore } from '$lib/stores/system-info.svelte.js';
import { settingsStore } from '$lib/stores/settings.svelte.js';
import { hostIntegrityFailed } from '$lib/stores/app';
import { evaluateHostname, isInsideTREGCCADDomain } from './hostIntegrity';

let checkPromise: Promise<boolean> | null = null;

/**
 * Runs the host integrity check exactly once per app session and resolves
 * only once `hostIntegrityFailed` reflects the result. The check itself is
 * async (it may need to fetch machine info from the backend), so callers
 * that gate behaviour on it — e.g. blocking a PEC opened via double-click
 * at startup, before the reactive check in the layout has had a chance to
 * run — must await this instead of reading the store synchronously.
 */
export function ensureHostIntegrityChecked(): Promise<boolean> {
  if (checkPromise) return checkPromise;

  checkPromise = (async () => {
    try {
      if (!systemInfoStore.data) {
        const info = await GetExtendedMachineData();
        systemInfoStore.setData(info);
      }

      if (!settingsStore.settings.enableHostIntegrityCheck) return false;

      const machineInfo = systemInfoStore.data;
      if (!machineInfo) return false;

      const hostnameOk = evaluateHostname(machineInfo.Hostname);
      const domainOk = isInsideTREGCCADDomain(machineInfo.ADDomain);
      const failed = !hostnameOk || !domainOk;
      hostIntegrityFailed.set(failed);
      return failed;
    } catch (e) {
      console.error('Host integrity check failed', e);
      return false;
    }
  })();

  return checkPromise;
}
