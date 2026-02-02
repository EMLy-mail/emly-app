<script lang="ts">
	import { Button } from "$lib/components/ui/button";
	import * as m from "$lib/paraglide/messages.js";

	let { onSave, onReset } = $props();
</script>

<div class="flex items-center gap-4 rounded-lg border bg-background px-4 py-3 shadow-lg w-full max-w-md">
	<span class="text-sm text-muted-foreground flex-1">
		{m.settings_unsaved_toast_message()}
	</span>

	<div class="ml-auto flex gap-2">
		<Button variant="ghost" onclick={onReset}>
			{m.settings_unsaved_toast_reset()}
		</Button>
		<Button onclick={onSave}>
			{m.settings_unsaved_toast_save()}
		</Button>
	</div>
</div>

<style>
	.bg-background {
		background-color: oklch(0.205 0 0);
	}
</style>
