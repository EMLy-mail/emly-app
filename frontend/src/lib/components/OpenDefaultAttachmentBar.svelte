<script lang="ts">
	import { Button } from "$lib/components/ui/button";
	import * as m from "$lib/paraglide/messages.js";
    import { FileQuestionMark } from "@lucide/svelte";

	let { onDownload, onCancel } = $props();

	/** How long the toast waits for an answer before asking for attention. */
	const NUDGE_DELAY = 5000;

	let nudging = $state(false);
	let hovered = $state(false);
	let focused = $state(false);

	/* The user is already looking at the bar, so there is nothing to nudge. */
	let engaged = $derived(hovered || focused);

	// Still unanswered after NUDGE_DELAY: blink the border like a turn signal
	// until the user picks one of the two buttons. Engaging the bar stops the
	// blink, and leaving it starts the wait again from zero - the effect tears
	// its timer down and builds a new one every time `engaged` flips.
	$effect(() => {
		if (engaged) {
			nudging = false;
			return;
		}
		const timer = setTimeout(() => (nudging = true), NUDGE_DELAY);
		return () => clearTimeout(timer);
	});

	/*
	 * Both handlers dismiss the toast, so the blink would stop with the
	 * component anyway - but the unmount is not instant and a button that keeps
	 * flashing after it was clicked reads as if the click missed.
	 */
	function answer(handler: () => void) {
		return () => {
			nudging = false;
			handler?.();
		};
	}
</script>

<!--
  focusin/focusout are the bubbling pair, so tabbing between the two buttons is
  seen here. Moving between them fires focusout before focusin, which would read
  as a moment of "left" - relatedTarget says where focus actually went.
-->
<div
	class="relative flex items-center gap-4 rounded-lg border bg-card px-4 py-3 shadow-lg w-full max-w-md"
	class:nudging
	role="group"
	aria-label={m.attachment_default_open_toast()}
	onpointerenter={() => (hovered = true)}
	onpointerleave={() => (hovered = false)}
	onfocusin={() => (focused = true)}
	onfocusout={(event) => {
		const next = event.relatedTarget;
		focused = next instanceof Node && event.currentTarget.contains(next);
	}}
>
	<FileQuestionMark class="shrink-0 text-white" />
	<span class="text-sm text-white flex-1">
		{m.attachment_default_open_toast()}
	</span>

	<div class="ml-auto flex gap-2">
		<Button variant="ghost" onclick={answer(onDownload)}>
			{m.attachment_default_open_toast_action()}
		</Button>
		<Button onclick={answer(onCancel)}>
			{m.attachment_default_open_toast_cancel()}
		</Button>
	</div>
</div>

<style>
	/*
	 * The glow rides on an overlay rather than on the bar itself: animating the
	 * real border and box-shadow would mean restating the card's own border
	 * colour and shadow-lg in every keyframe, and they would drift apart the
	 * next time the theme moves. Only opacity animates here.
	 */
	.nudging::after {
		content: "";
		position: absolute;
		inset: 0;
		border-radius: inherit;
		border: 1px solid rgb(255 255 255 / 0.9);
		box-shadow:
			0 0 0 1px rgb(255 255 255 / 0.35),
			0 0 18px 2px rgb(255 255 255 / 0.45),
			inset 0 0 12px rgb(255 255 255 / 0.18);
		pointer-events: none;
		/* A car indicator is a hard on/off, not a fade: keep the steps square. */
		animation: turn-signal 0.85s steps(1, end) infinite;
	}

	@keyframes turn-signal {
		0%,
		50% {
			opacity: 1;
		}
		50.01%,
		100% {
			opacity: 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.nudging::after {
			animation: none;
			opacity: 1;
		}
	}
</style>
