<script lang="ts">
	import { onMount } from 'svelte';

	let { children } = $props();

	onMount(() => {
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
