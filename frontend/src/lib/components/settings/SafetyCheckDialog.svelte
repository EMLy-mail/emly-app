<script lang="ts">
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import { Badge } from "$lib/components/ui/badge/index.js";
    import { Button, buttonVariants } from "$lib/components/ui/button/index.js";
    import { Separator } from "$lib/components/ui/separator";
    import {
        CircleCheck,
        CircleX,
        Loader2,
        OctagonX,
        TriangleAlert,
    } from "@lucide/svelte";
    import * as m from "$lib/paraglide/messages";
    import { LogDebug } from "$lib/wailsjs/runtime/runtime";
    import {
        GetUpdaterADStatus,
        GetUpdaterSystemInfo,
    } from "$lib/wailsjs/go/main/App";
    import { settingsStore } from "$lib/stores/settings.svelte";
    import { runningInDebugMode } from "$lib/stores/app";
    import { systemInfoStore } from "$lib/stores/system-info.svelte.js";
    import { updaterStatusStore } from "$lib/stores/updater-status.svelte.js";
    import {
        evaluateHostname,
        isInsideTREGCCADDomain,
    } from "$lib/utils/hostIntegrity";
    import { logIPCRequest, logIPCResponse, logIPCError } from "$lib/utils/ipcLog";

    let { open = $bindable(false) }: { open: boolean } = $props();

    let machineData = $derived(systemInfoStore.data);

    // "Run safety check" — a one-shot diagnostic combining host identity
    // (hostname/AD domain) with the EMLy Updater's local + IPC status, plus
    // a cross-check of AD domain/hostname as seen by the IPC-connected
    // SYSTEM-privileged service vs. this (unprivileged) process, as a
    // double failsafe against a spoofed/compromised local view.
    type SafetyCheckStanding = "perfect" | "usable" | "limited";

    type SafetyCheckResult = {
        hostnameOk: boolean;
        adDomainOk: boolean;
        updaterInstalled: boolean;
        updaterRunning: boolean;
        ipcInstalled: boolean;
        ipcResponding: boolean;
        crossCheckOk: boolean;
        hostIntegrityToggleOk: boolean;
        standing: SafetyCheckStanding;
    };

    let runningSafetyCheck = $state(false);
    let safetyCheckResult = $state<SafetyCheckResult | null>(null);

    async function runSafetyCheck() {
        runningSafetyCheck = true;
        try {
            await updaterStatusStore.refreshUpdaterStatus();
            await updaterStatusStore.checkUpdaterIPCStatus();

            const hostname = machineData?.Hostname ?? "";
            const adDomain = machineData?.ADDomain ?? "";

            const hostnameOk = evaluateHostname(hostname);
            const adDomainOk = isInsideTREGCCADDomain(adDomain);

            const updaterInstalled = !!updaterStatusStore.installed;
            const updaterRunning = !!updaterStatusStore.running;
            const ipcInstalled = !!updaterStatusStore.ipcActive;
            const ipcResponding = !!updaterStatusStore.ipcValid;

            let crossCheckOk = false;
            if (ipcInstalled && ipcResponding) {
                try {
                    logIPCRequest("GetUpdaterADStatus");
                    logIPCRequest("GetUpdaterSystemInfo");
                    const [adStatus, sysInfo] = await Promise.all([
                        GetUpdaterADStatus(),
                        GetUpdaterSystemInfo(),
                    ]);
                    logIPCResponse("GetUpdaterADStatus", adStatus);
                    logIPCResponse("GetUpdaterSystemInfo", sysInfo);
                    crossCheckOk =
                        adStatus.ADDomain === adDomain &&
                        sysInfo.Hostname === hostname;
                } catch (err) {
                    logIPCError("GetUpdaterADStatus/GetUpdaterSystemInfo", err);
                    LogDebug(`Safety check IPC cross-check failed: ${err}`);
                    crossCheckOk = false;
                }
            }

            // The Host Integrity toggle must be on — the only exemption is
            // an actual DEBUG MODE build ($runningInDebugMode, the Go-side
            // build flag), never merely running under `wails dev`/vite
            // (`dev`) and never a recognized dev machine (which only
            // bypasses the hostname/AD checks above). Disabling it outside
            // of a debug build forces a Limited standing regardless of
            // every other result.
            const isDebugBuild = $runningInDebugMode;
            const hostIntegrityToggleOk =
                settingsStore.settings.enableHostIntegrityCheck ||
                isDebugBuild;

            let standing: SafetyCheckStanding;
            if (!hostIntegrityToggleOk) {
                standing = "limited";
            } else if (
                hostnameOk &&
                adDomainOk &&
                updaterInstalled &&
                updaterRunning &&
                ipcInstalled &&
                ipcResponding &&
                crossCheckOk
            ) {
                standing = "perfect";
            } else if (
                hostnameOk &&
                adDomainOk &&
                updaterInstalled &&
                !updaterRunning
            ) {
                standing = "usable";
            } else {
                standing = "limited";
            }

            safetyCheckResult = {
                hostnameOk,
                adDomainOk,
                updaterInstalled,
                updaterRunning,
                ipcInstalled,
                ipcResponding,
                crossCheckOk,
                hostIntegrityToggleOk,
                standing,
            };
        } finally {
            runningSafetyCheck = false;
        }
    }

    $effect(() => {
        if (open) {
            runSafetyCheck();
        }
    });
</script>

<Dialog.Root bind:open>
    <Dialog.Content class="sm:max-w-md">
        <Dialog.Header>
            <Dialog.Title>{m.settings_safety_check_dialog_title()}</Dialog.Title>
            <Dialog.Description>
                {m.settings_safety_check_dialog_description()}
            </Dialog.Description>
        </Dialog.Header>

        {#if runningSafetyCheck && !safetyCheckResult}
            <div class="flex items-center justify-center gap-2 py-6 text-sm text-muted-foreground">
                <Loader2 class="size-4 animate-spin" />
                {m.settings_safety_check_running()}
            </div>
        {:else if safetyCheckResult}
            <div class="space-y-2">
                <div class="flex items-center justify-between gap-2 text-sm">
                    {m.settings_safety_check_item_hostname()}
                    {#if safetyCheckResult.hostnameOk}
                        <CircleCheck class="size-4 text-green-500" />
                    {:else}
                        <CircleX class="size-4 text-red-500" />
                    {/if}
                </div>
                <div class="flex items-center justify-between gap-2 text-sm">
                    {m.settings_safety_check_item_ad_domain()}
                    {#if safetyCheckResult.adDomainOk}
                        <CircleCheck class="size-4 text-green-500" />
                    {:else}
                        <CircleX class="size-4 text-red-500" />
                    {/if}
                </div>
                <div class="flex items-center justify-between gap-2 text-sm">
                    {m.settings_safety_check_item_updater_installed()}
                    {#if safetyCheckResult.updaterInstalled}
                        <CircleCheck class="size-4 text-green-500" />
                    {:else}
                        <CircleX class="size-4 text-red-500" />
                    {/if}
                </div>
                <div class="flex items-center justify-between gap-2 text-sm">
                    {m.settings_safety_check_item_updater_running()}
                    {#if safetyCheckResult.updaterRunning}
                        <CircleCheck class="size-4 text-green-500" />
                    {:else}
                        <CircleX class="size-4 text-red-500" />
                    {/if}
                </div>
                <div class="flex items-center justify-between gap-2 text-sm">
                    {m.settings_safety_check_item_ipc_installed()}
                    {#if safetyCheckResult.ipcInstalled}
                        <CircleCheck class="size-4 text-green-500" />
                    {:else}
                        <CircleX class="size-4 text-red-500" />
                    {/if}
                </div>
                <div class="flex items-center justify-between gap-2 text-sm">
                    {m.settings_safety_check_item_ipc_responding()}
                    {#if safetyCheckResult.ipcResponding}
                        <CircleCheck class="size-4 text-green-500" />
                    {:else}
                        <CircleX class="size-4 text-red-500" />
                    {/if}
                </div>
                <div class="flex items-center justify-between gap-2 text-sm">
                    {m.settings_safety_check_item_cross_check()}
                    {#if safetyCheckResult.crossCheckOk}
                        <CircleCheck class="size-4 text-green-500" />
                    {:else}
                        <CircleX class="size-4 text-red-500" />
                    {/if}
                </div>
                <div class="flex items-center justify-between gap-2 text-sm">
                    {m.settings_safety_check_item_host_integrity_toggle()}
                    {#if safetyCheckResult.hostIntegrityToggleOk}
                        <CircleCheck class="size-4 text-green-500" />
                    {:else}
                        <CircleX class="size-4 text-red-500" />
                    {/if}
                </div>

                <Separator />

                <div class="flex items-center justify-between gap-2 pt-1">
                    <span class="text-sm font-medium">
                        {m.settings_safety_check_standing_label()}
                    </span>
                    {#if safetyCheckResult.standing === "perfect"}
                        <Badge class="bg-green-600 text-white border-transparent">
                            <CircleCheck />
                            {m.settings_safety_check_standing_perfect()}
                        </Badge>
                    {:else if safetyCheckResult.standing === "usable"}
                        <Badge class="bg-yellow-500 text-black border-transparent">
                            <TriangleAlert />
                            {m.settings_safety_check_standing_usable()}
                        </Badge>
                    {:else}
                        <Badge variant="destructive">
                            <OctagonX />
                            {m.settings_safety_check_standing_limited()}
                        </Badge>
                    {/if}
                </div>
                <div class="text-xs text-muted-foreground">
                    {#if safetyCheckResult.standing === "perfect"}
                        {m.settings_safety_check_standing_perfect_hint()}
                    {:else if safetyCheckResult.standing === "usable"}
                        {m.settings_safety_check_standing_usable_hint()}
                    {:else}
                        {m.settings_safety_check_standing_limited_hint()}
                    {/if}
                </div>
            </div>
        {/if}

        <Dialog.Footer>
            <Button
                variant="outline"
                class="cursor-pointer hover:cursor-pointer"
                onclick={runSafetyCheck}
                disabled={runningSafetyCheck}
            >
                {#if runningSafetyCheck}
                    <Loader2 class="size-4 animate-spin" />
                {/if}
                {m.settings_safety_check_rerun_btn()}
            </Button>
            <Dialog.Close
                class={`${buttonVariants({ variant: "destructive" })} cursor-pointer hover:cursor-pointer`}
            >
                {m.settings_safety_check_close_btn()}
            </Dialog.Close>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
