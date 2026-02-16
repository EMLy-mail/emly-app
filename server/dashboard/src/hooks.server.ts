import type { Handle } from '@sveltejs/kit';
import { lucia } from '$lib/server/auth';
import { initLogger, Log } from '$lib/server/logger';

// Initialize dashboard logger
initLogger();

export const handle: Handle = async ({ event, resolve }) => {
	const ip =
		event.request.headers.get('x-forwarded-for')?.split(',')[0]?.trim() ||
		event.request.headers.get('x-real-ip') ||
		event.getClientAddress?.() ||
		'unknown';
	Log('HTTP', `${event.request.method} ${event.url.pathname} from ${ip}`);

	const sessionId = event.cookies.get(lucia.sessionCookieName);

	if (!sessionId) {
		event.locals.user = null;
		event.locals.session = null;
		return resolve(event);
	}

	const { session, user } = await lucia.validateSession(sessionId);

	if (session && session.fresh) {
		const sessionCookie = lucia.createSessionCookie(session.id);
		event.cookies.set(sessionCookie.name, sessionCookie.value, {
			path: '.',
			...sessionCookie.attributes
		});
	}

	if (!session) {
		Log('AUTH', `Invalid session from ip=${ip}`);
		const sessionCookie = lucia.createBlankSessionCookie();
		event.cookies.set(sessionCookie.name, sessionCookie.value, {
			path: '.',
			...sessionCookie.attributes
		});
	}

	// If user is disabled, invalidate their session and clear cookie
	if (session && user && !user.enabled) {
		Log('AUTH', `Disabled user rejected: username=${user.username} ip=${ip}`);
		await lucia.invalidateSession(session.id);
		const sessionCookie = lucia.createBlankSessionCookie();
		event.cookies.set(sessionCookie.name, sessionCookie.value, {
			path: '.',
			...sessionCookie.attributes
		});
		event.locals.user = null;
		event.locals.session = null;
		return resolve(event);
	}

	event.locals.user = user;
	event.locals.session = session;
	return resolve(event);
};
