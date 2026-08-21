<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { setupConsoleLogger } from '$lib/utils/logger-hook';
	import { ensureHostIntegrityChecked } from '$lib/utils/hostIntegrityCheck';
	import { trace } from '$lib/utils/startupTrace';
	import "./layout.css";
	
	let { children } = $props();

	// WebView2 opens a new native window when a link is Shift+clicked. Block
	// that at the source (capture phase, before it reaches the anchor's
	// default action) so it never spawns unmanaged windows outside the app.
	function suppressShiftClickNewWindow(e: MouseEvent) {
		if (!e.shiftKey && !e.ctrlKey && !e.altKey) return;
		if ((e.target as HTMLElement)?.closest('a[href]')) {
			e.preventDefault();
		}
	}

	onMount(async () => {
		trace('fe_layout_mount_start');
		document.addEventListener('click', suppressShiftClickNewWindow, true);

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

		// Fase 1 – Recupero dati macchina + verifica integrità host
		if (stepEl) stepEl.textContent = t('Recupero dati...', 'Fetching data...');
		trace('fe_layout_host_integrity_start');
		await ensureHostIntegrityChecked();
		trace('fe_layout_host_integrity_done');

		// Fase 2 – Caricamento layout
		if (stepEl) stepEl.textContent = t('Caricamento layout...', 'Loading layout...');
		await tick();
		trace('fe_layout_tick_done');

		if (loader) {
			loader.style.opacity = '0';
			setTimeout(() => loader.remove(), 300);
		}
		trace('fe_loader_removed');
	});
</script>

{@render children()}
