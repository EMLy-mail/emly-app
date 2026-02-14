import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { db } from '$lib/server/db';
import { bugReports } from '$lib/schema';
import { eq } from 'drizzle-orm';

export const PATCH: RequestHandler = async ({ params, request }) => {
	const id = Number(params.id);
	if (isNaN(id)) throw error(400, 'Invalid report ID');

	const body = await request.json();
	const { status } = body;

	if (!['new', 'in_review', 'resolved', 'closed'].includes(status)) {
		throw error(400, 'Invalid status');
	}

	const [result] = await db
		.update(bugReports)
		.set({ status })
		.where(eq(bugReports.id, id));

	if (result.affectedRows === 0) throw error(404, 'Report not found');

	return json({ success: true });
};

export const DELETE: RequestHandler = async ({ params }) => {
	const id = Number(params.id);
	if (isNaN(id)) throw error(400, 'Invalid report ID');

	const [result] = await db
		.delete(bugReports)
		.where(eq(bugReports.id, id));

	if (result.affectedRows === 0) throw error(404, 'Report not found');

	return json({ success: true });
};
