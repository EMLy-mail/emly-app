<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { setupConsoleLogger } from '$lib/utils/logger-hook';
	import { GetMachineData } from '$lib/wailsjs/go/main/App';
	import { systemInfoStore } from '$lib/stores/system-info.svelte.js';
	import "./layout.css";
	
	let { children } = $props();

	onMount(async () => {
		setupConsoleLogger();

		const loader = document.getElementById('app-loading');
		const stepEl = document.getElementById('loading-step');
		const loadingEl = document.getElementById('loading-text');

		let lang = 'en';
		try {
			const s = JSON.parse(localStorage.getItem('emly_gui_settings') || '{}');
			lang = s.selectedLanguage || 'en';
		} catch { /* fallback a 'en' */ }

		const t = (it: string, en: string) => lang === 'it' ? it : en;

		if (loadingEl) loadingEl.textContent = t('Caricamento...', 'Loading...');

		// Clear the pre-Svelte crash hint timeout now that JS has loaded
		clearTimeout((window as any).__emlyLoadTimeout);

		// Fase 1 – Recupero dati macchina
		if (stepEl) stepEl.textContent = t('Recupero dati...', 'Fetching data...');
		try {
			if(systemInfoStore.data) {
				return;
			}
			const info = await GetMachineData();
			systemInfoStore.setData(info);
		} catch (e) {
			console.error('Failed to fetch machine data', e);
		}

		// Fase 2 – Caricamento layout
		if (stepEl) stepEl.textContent = t('Caricamento layout...', 'Loading layout...');
		await tick();

		if (loader) {
			loader.style.opacity = '0';
			setTimeout(() => loader.remove(), 300);
		}
	});
</script>

{@render children()}
