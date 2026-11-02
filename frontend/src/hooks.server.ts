import { sequence, type Handle } from '@sveltejs/kit/hooks';
import { paraglideMiddleware } from '#lib/paraglide/server.js';

// event.request is read-only in SvelteKit 3, so the de-localized request
// paraglide hands back can't be swapped in. It's the original request anyway:
// the locale strategy (cookie, globalVariable, baseLocale) has no "url" step.
const handleParaglide: Handle = ({ event, resolve }) => paraglideMiddleware(event.request, ({ locale }) =>
	resolve(event, {
		transformPageChunk: ({ html }) => html.replace('%paraglide.lang%', locale)
	})
);

export const handle: Handle = sequence(handleParaglide);
