<script lang="ts">
    import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
    import type { SettingsResetEntry } from "$lib/stores/settings-reset.svelte";
    import * as m from "$lib/paraglide/messages";

    let {
        open = $bindable(false),
        entries,
        onAcknowledge,
    }: {
        open: boolean;
        entries: SettingsResetEntry[];
        onAcknowledge: () => void;
    } = $props();
</script>

<AlertDialog.Root bind:open>
    <AlertDialog.Content>
        <AlertDialog.Header>
            <AlertDialog.Title>{m.settings_reset_dialog_title()}</AlertDialog.Title>
            <AlertDialog.Description>
                {m.settings_reset_dialog_description()}
            </AlertDialog.Description>
        </AlertDialog.Header>
        <ul class="reset-list">
            {#each entries as entry (entry.id)}
                <li class="reset-entry">
                    <span class="reset-setting">{entry.setting}</span>
                    <span class="reset-reason">{entry.reason}</span>
                </li>
            {/each}
        </ul>
        <AlertDialog.Footer>
            <AlertDialog.Action onclick={onAcknowledge}
                >{m.settings_reset_dialog_confirm()}</AlertDialog.Action
            >
        </AlertDialog.Footer>
    </AlertDialog.Content>
</AlertDialog.Root>

<style>
    .reset-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
        margin: 4px 0;
        padding: 0;
        list-style: none;
        max-height: 40vh;
        overflow-y: auto;
    }

    .reset-entry {
        display: flex;
        flex-direction: column;
        gap: 2px;
        background: var(--muted);
        border: 1px solid var(--border);
        border-radius: 8px;
        padding: 10px 12px;
    }

    .reset-setting {
        font-size: 11px;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        color: var(--muted-foreground);
    }

    .reset-reason {
        font-size: 12px;
        color: var(--foreground);
    }
</style>
