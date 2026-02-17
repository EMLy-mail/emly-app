import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { bugReports } from '$lib/schema';
import { eq, count } from 'drizzle-orm';

export const load: LayoutServerLoad = async ({ locals, url }) => {
	if (url.pathname === '/login') {
		return { newCount: 0, user: null };
	}

	if (!locals.user) {
		redirect(302, '/login');
	}

	const [result] = await db
		.select({ count: count() })
		.from(bugReports)
		.where(eq(bugReports.status, 'new'));

	return {
		newCount: result.count,
		user: locals.user
	};
};
