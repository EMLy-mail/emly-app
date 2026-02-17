import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { presenceMap, broadcastPresence } from '../state';

export const POST: RequestHandler = async ({ request, locals }) => {
	if (!locals.user) {
		error(401, 'Unauthorized');
	}

	const body = await request.json();
	const { currentPath, reportId } = body;

	presenceMap.set(locals.user.id, {
		userId: locals.user.id,
		username: locals.user.username,
		displayname: locals.user.displayname,
		currentPath: currentPath || '/',
		reportId: reportId || null,
		lastSeen: Date.now()
	});

	broadcastPresence();

	return json({ ok: true });
};
