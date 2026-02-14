import type { PageServerLoad } from './$types';
import { db } from '$lib/server/db';
import { bugReports, bugReportFiles } from '$lib/schema';
import { eq, like, or, count, sql, desc } from 'drizzle-orm';

export const load: PageServerLoad = async ({ url }) => {
	const page = Math.max(1, Number(url.searchParams.get('page')) || 1);
	const pageSize = Math.min(50, Math.max(10, Number(url.searchParams.get('pageSize')) || 20));
	const status = url.searchParams.get('status') || '';
	const search = url.searchParams.get('search') || '';

	const conditions = [];
	if (status && ['new', 'in_review', 'resolved', 'closed'].includes(status)) {
		conditions.push(eq(bugReports.status, status as 'new' | 'in_review' | 'resolved' | 'closed'));
	}
	if (search) {
		const pattern = `%${search}%`;
		conditions.push(
			or(
				like(bugReports.hostname, pattern),
				like(bugReports.os_user, pattern),
				like(bugReports.name, pattern),
				like(bugReports.email, pattern)
			)!
		);
	}

	const where = conditions.length > 0
		? conditions.length === 1
			? conditions[0]
			: sql`${conditions[0]} AND ${conditions[1]}`
		: undefined;

	const [totalResult] = await db
		.select({ count: count() })
		.from(bugReports)
		.where(where);

	const total = totalResult.count;
	const totalPages = Math.max(1, Math.ceil(total / pageSize));

	const fileCountSubquery = db
		.select({
			report_id: bugReportFiles.report_id,
			file_count: count().as('file_count')
		})
		.from(bugReportFiles)
		.groupBy(bugReportFiles.report_id)
		.as('fc');

	const reports = await db
		.select({
			id: bugReports.id,
			name: bugReports.name,
			email: bugReports.email,
			hostname: bugReports.hostname,
			os_user: bugReports.os_user,
			status: bugReports.status,
			created_at: bugReports.created_at,
			file_count: sql<number>`COALESCE(${fileCountSubquery.file_count}, 0)`.as('file_count')
		})
		.from(bugReports)
		.leftJoin(fileCountSubquery, eq(bugReports.id, fileCountSubquery.report_id))
		.where(where)
		.orderBy(desc(bugReports.created_at))
		.limit(pageSize)
		.offset((page - 1) * pageSize);

	return {
		reports: reports.map((r) => ({
			...r,
			created_at: r.created_at.toISOString()
		})),
		pagination: { page, pageSize, total, totalPages },
		filters: { status, search }
	};
};
