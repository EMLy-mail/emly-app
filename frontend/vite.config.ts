import devtoolsJson from 'vite-plugin-devtools-json';
import { paraglideVitePlugin } from '@inlang/paraglide-js';
import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, type Plugin } from 'vite';

// @embedpdf pre-compiled Svelte components pass arrays to rest_props(),
// but Svelte 5.56+ expects a Set. This plugin patches the files at build time.
// Handles both original dist paths ($.rest_props) and Vite pre-bundled paths (rest_props).
function patchEmbedPdfRestProps(): Plugin {
	return {
		name: 'patch-embedpdf-rest-props',
		transform(code, id) {
			if (!id.includes('@embedpdf')) return;
			if (!code.includes('rest_props')) return;
			// Match both "$.rest_props(x, [...])" and "rest_props(x, [...])"
			const patched = code.replace(
				/([\w$]*\.?rest_props\([^,]+,\s*)\[([^\]]*)\]/gs,
				'$1new Set([$2])'
			);
			if (patched !== code) return { code: patched, map: null };
		}
	};
}

export default defineConfig({
	plugins: [
		patchEmbedPdfRestProps(),
		tailwindcss(),
		sveltekit(),
		paraglideVitePlugin({ project: './project.inlang', outdir: './src/lib/paraglide'}),
		devtoolsJson()
	],
	optimizeDeps: {
		// Pre-bundle @embedpdf eagerly at server start to avoid 504 timeouts during serving
		include: [
			'@embedpdf/core',
			'@embedpdf/core/svelte',
			'@embedpdf/engines/svelte',
			'@embedpdf/plugin-viewport/svelte',
			'@embedpdf/plugin-scroll/svelte',
			'@embedpdf/plugin-document-manager/svelte',
			'@embedpdf/plugin-render/svelte',
			'@embedpdf/plugin-zoom/svelte',
			'@embedpdf/plugin-rotate/svelte',
		],
		esbuildOptions: {
			target: 'es2022',
		},
	},
	esbuild: {
		target: 'es2022',
	},
	build: {
		target: 'es2022',
	}
});
