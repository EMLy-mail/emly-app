<script lang="ts">
	import { Button } from "$lib/components/ui/button";
	import * as m from "$lib/paraglide/messages.js";
    import { FileQuestionMark } from "@lucide/svelte";

	let { onDownload, onCancel } = $props();
</script>

<div class="flex items-center gap-4 rounded-lg border bg-card px-4 py-3 shadow-lg w-full max-w-md">
	<FileQuestionMark class="shrink-0 text-muted-foreground" />
	<span class="text-sm text-muted-foreground flex-1">
		{m.attachment_default_open_toast()}
	</span>

	<div class="ml-auto flex gap-2">
		<Button variant="ghost" onclick={onDownload}>
			{m.attachment_default_open_toast_action()}
		</Button>
		<Button onclick={onCancel}>
			{m.attachment_default_open_toast_cancel()}
		</Button>
	</div>
</div>
