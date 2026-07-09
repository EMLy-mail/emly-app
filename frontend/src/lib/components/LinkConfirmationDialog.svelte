<script lang="ts">
    import * as AlertDialog from "$lib/components/ui/alert-dialog/index.js";
    import * as m from "$lib/paraglide/messages";

    let {
        open = $bindable(false),
        url,
        onConfirm,
        onCancel,
    }: {
        open: boolean;
        url: string;
        onConfirm: () => void;
        onCancel: () => void;
    } = $props();
</script>

<AlertDialog.Root bind:open>
    <AlertDialog.Content>
        <AlertDialog.Header>
            <AlertDialog.Title>{m.mail_link_security_title()}</AlertDialog.Title
            >
            <AlertDialog.Description>
                {m.mail_link_security_description()}
            </AlertDialog.Description>
        </AlertDialog.Header>
        <div class="link-url-box">
            <span class="link-url-label"
                >{m.mail_link_security_url_label()}</span
            >
            <span class="link-url-value">{url}</span>
        </div>
        <AlertDialog.Footer>
            <AlertDialog.Cancel onclick={onCancel}
                >{m.mail_link_security_cancel()}</AlertDialog.Cancel
            >
            <AlertDialog.Action onclick={onConfirm}
                >{m.mail_link_security_open()}</AlertDialog.Action
            >
        </AlertDialog.Footer>
    </AlertDialog.Content>
</AlertDialog.Root>

<style>
    .link-url-box {
        display: flex;
        flex-direction: column;
        gap: 4px;
        background: var(--muted);
        border: 1px solid var(--border);
        border-radius: 8px;
        padding: 10px 12px;
        margin: 4px 0;
    }

    .link-url-label {
        font-size: 11px;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        color: var(--muted-foreground);
    }

    .link-url-value {
        font-size: 12px;
        color: var(--foreground);
        word-break: break-all;
        font-family: monospace;
    }
</style>
