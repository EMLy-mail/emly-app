import devtoolsJson from 'vite-plugin-devtools-json';
import { paraglideVitePlugin } from '@inlang/paraglide-js';
import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		paraglideVitePlugin({ project: './project.inlang', outdir: './src/lib/paraglide'}),
		devtoolsJson()
	],
	optimizeDeps: {
		// Pre-bundle eagerly at server start to avoid 504 timeouts during serving
		include: [
			'pdfjs-dist',
			'@lucide/svelte',
			'@lucide/svelte/icons',
			'@lucide/svelte/icons/circle-check',
			'@lucide/svelte/icons/info',
			'@lucide/svelte/icons/loader-2',
			'@lucide/svelte/icons/octagon-x',
			'@lucide/svelte/icons/panel-left',
			'@lucide/svelte/icons/settings',
			'@lucide/svelte/icons/triangle-alert',
			'@lucide/svelte/icons/check',
			'@lucide/svelte/icons/circle', 
			'@lucide/svelte/icons/minus', 
			'svelte-flags',
			'@lucide/svelte/icons/x',
			'bits-ui',
			'dompurify',
			'mode-watcher',
			'svelte-sonner',
			'tailwind-merge',
			'tailwind-variants',
			'@lucide/svelte/icons/chevron-right',
			'pdf-lib',
			'@lucide/svelte/icons/chevron-down', 
			'@lucide/svelte/icons/chevron-up',
		],
	},
	build: {
		target: 'es2022',
	}
});
