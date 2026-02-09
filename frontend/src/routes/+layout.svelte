<script lang="ts">
	import { onMount } from 'svelte';
	import { setupConsoleLogger } from '$lib/utils/logger-hook';
	import "./layout.css";

	let { children } = $props();

	onMount(() => {
		setupConsoleLogger();
		const loader = document.getElementById('app-loading');
		if (loader) {
			loader.style.opacity = '0';
			setTimeout(() => {
				loader.remove();
			}, 300);
		}
	});
</script>

{@render children()}
