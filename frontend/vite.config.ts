import { createRequire } from 'node:module';
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { dirname, join, relative, sep } from 'node:path';
import devtoolsJson from 'vite-plugin-devtools-json';
import { paraglideVitePlugin } from '@inlang/paraglide-js';
import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, type Plugin } from 'vite';

/*
 * pdf.js fetches part of itself at runtime: the JBIG2 and JPEG2000 decoders
 * and the colour engine are .wasm blobs, and CJK cmaps, the standard fonts and
 * the default ICC profile are data files. Nothing ever imports them, so no
 * bundler can see them, and pdf.js only reaches them if it is handed a base
 * URL per group. Without that a scanned fax PDF fails with "JBig2 failed to
 * initialize" and the page renders blank.
 *
 * Serving them from node_modules in dev and emitting them verbatim in the
 * build keeps the files at the version pdfjs-dist is pinned to, where a copy
 * under static/ would silently rot on the next upgrade. The URL prefix is
 * PDFJS_ASSET_BASE in PdfjsViewer.svelte - the two have to agree.
 */
const PDFJS_ASSET_DIRS = ['wasm', 'cmaps', 'standard_fonts', 'iccs'];
const PDFJS_ASSET_PREFIX = 'pdfjs';

const MIME_TYPES: Record<string, string> = {
	'.wasm': 'application/wasm',
	'.js': 'text/javascript',
	'.mjs': 'text/javascript',
	'.bcmap': 'application/octet-stream',
	'.pfb': 'application/octet-stream',
	'.icc': 'application/vnd.iccprofile'
};

function pdfjsAssets(): Plugin {
	const require = createRequire(import.meta.url);
	const root = dirname(require.resolve('pdfjs-dist/package.json'));

	/** Every asset file, keyed by the URL path it is served at. */
	const files = new Map<string, string>();
	for (const dir of PDFJS_ASSET_DIRS) {
		const base = join(root, dir);
		const walk = (current: string) => {
			for (const entry of readdirSync(current)) {
				const absolute = join(current, entry);
				if (statSync(absolute).isDirectory()) {
					walk(absolute);
					continue;
				}
				const url = relative(root, absolute).split(sep).join('/');
				files.set(`${PDFJS_ASSET_PREFIX}/${url}`, absolute);
			}
		};
		walk(base);
	}

	let ssr = false;

	return {
		name: 'emly-pdfjs-assets',

		configResolved(config) {
			ssr = !!config.build.ssr;
		},

		configureServer(server) {
			server.middlewares.use((req, res, next) => {
				const path = req.url?.split('?')[0].replace(/^\//, '') ?? '';
				const absolute = files.get(decodeURIComponent(path));
				if (!absolute) return next();

				const extension = absolute.slice(absolute.lastIndexOf('.'));
				res.setHeader('Content-Type', MIME_TYPES[extension] ?? 'application/octet-stream');
				res.end(readFileSync(absolute));
			});
		},

		// The client bundle only. SvelteKit also runs a server build, and the
		// adapter ships the client one.
		generateBundle() {
			if (ssr) return;
			for (const [fileName, absolute] of files) {
				this.emitFile({ type: 'asset', fileName, source: readFileSync(absolute) });
			}
		}
	};
}

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		paraglideVitePlugin({ project: './project.inlang', outdir: './src/lib/paraglide'}),
		devtoolsJson(),
		pdfjsAssets()
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
