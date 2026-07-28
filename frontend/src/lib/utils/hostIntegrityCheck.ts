import {
  GetExtendedMachineData,
  GetUpdaterADStatus,
  GetUpdaterSystemInfo,
} from '$lib/wailsjs/go/main/App';
import { systemInfoStore } from '$lib/stores/system-info.svelte.js';
import { settingsStore } from '$lib/stores/settings.svelte.js';
import { hostIntegrityFailed, hostIntegrityStanding } from '$lib/stores/app';
import { updaterStatusStore } from '$lib/stores/updater-status.svelte.js';
import { evaluateHostname, isInsideTREGCCADDomain, deriveHostIntegrityStanding } from './hostIntegrity';
import { logIPCRequest, logIPCResponse, logIPCError, logIPCHandshake } from './ipcLog';
import { LogDebug } from '$lib/wailsjs/runtime/runtime';

let checkPromise: Promise<boolean> | null = null;
let standingPromise: Promise<void> | null = null;

/**
 * Runs the host integrity check exactly once per app session and resolves
 * only once `hostIntegrityFailed` reflects the result. The check itself is
 * async (it may need to fetch machine info from the backend), so callers
 * that gate behaviour on it — e.g. blocking a PEC opened via double-click
 * at startup, before the reactive check in the layout has had a chance to
 * run — must await this instead of reading the store synchronously.
 *
 * This only covers the fast path (hostname/AD domain/toggle) so callers
 * that await it (app mount, startup page load) are never blocked on the
 * slower Updater/IPC round trips. Once the fast path passes, a background
 * continuation (see resolveFullStanding) refines the result into the full
 * tri-state standing (perfect/acceptable/limited) and updates the stores
 * again if needed — callers reacting to the stores stay eventually
 * consistent even though this promise already resolved.
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
      hostIntegrityStanding.set(failed ? 'limited' : null);

      if (!failed) {
        resolveFullStanding(machineInfo.Hostname, machineInfo.ADDomain);
      }

      return failed;
    } catch (e) {
      console.error('Host integrity check failed', e);
      return false;
    }
  })();

  return checkPromise;
}

// Refines the fast-path result (hostname/AD/toggle only) into the full
// tri-state standing by also checking the EMLy Updater's local + IPC status,
// plus a cross-check of AD domain/hostname as seen by the IPC-connected
// SYSTEM-privileged service vs. this (unprivileged) process. Runs once per
// session, in the background, without blocking anything awaiting
// ensureHostIntegrityChecked().
function resolveFullStanding(hostname: string, adDomain: string): Promise<void> {
  if (standingPromise) return standingPromise;

  standingPromise = (async () => {
    try {
      await updaterStatusStore.refreshUpdaterStatus();
      await updaterStatusStore.checkUpdaterIPCStatus();

      const updaterInstalled = !!updaterStatusStore.installed;
      const updaterRunning = !!updaterStatusStore.running;
      const ipcActive = !!updaterStatusStore.ipcActive;
      const ipcResponding = !!updaterStatusStore.ipcValid;

      let crossCheckOk = false;
      if (ipcActive && ipcResponding) {
        try {
          logIPCRequest('GetUpdaterADStatus');
          logIPCRequest('GetUpdaterSystemInfo');
          const [adStatus, sysInfo] = await Promise.all([
            GetUpdaterADStatus(),
            GetUpdaterSystemInfo(),
          ]);
          logIPCResponse('GetUpdaterADStatus', adStatus);
          logIPCResponse('GetUpdaterSystemInfo', sysInfo);
          logIPCHandshake('GetUpdaterADStatus', adStatus);
          logIPCHandshake('GetUpdaterSystemInfo', sysInfo);
          crossCheckOk =
            adStatus.ADDomain === adDomain && sysInfo.Hostname === hostname;
        } catch (err) {
          logIPCError('GetUpdaterADStatus/GetUpdaterSystemInfo', err);
          LogDebug(`Host integrity IPC cross-check failed: ${err}`);
          crossCheckOk = false;
        }
      }

      const standing = deriveHostIntegrityStanding({
        hostnameOk: true,
        adDomainOk: true,
        hostIntegrityToggleOk: true,
        updaterInstalled,
        updaterRunning,
        ipcActive,
        ipcResponding,
        crossCheckOk,
      });

      hostIntegrityStanding.set(standing);
      if (standing === 'limited') {
        hostIntegrityFailed.set(true);
      }
    } catch (e) {
      console.error('Host integrity standing check failed', e);
    }
  })();

  return standingPromise;
}
