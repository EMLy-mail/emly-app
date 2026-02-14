import type { LayoutServerLoad } from './$types';
import { db } from '$lib/server/db';
import { bugReports } from '$lib/schema';
import { eq, count } from 'drizzle-orm';

export const load: LayoutServerLoad = async () => {
	const [result] = await db
		.select({ count: count() })
		.from(bugReports)
		.where(eq(bugReports.status, 'new'));

	return {
		newCount: result.count
	};
};
