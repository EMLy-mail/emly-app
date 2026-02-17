import type { PageServerLoad } from './$types';
import { db } from '$lib/server/db';
import { bugReports, bugReportFiles } from '$lib/schema';
import { eq, like, or, count, sql, desc, and } from 'drizzle-orm';

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
		conditions.push(
			or(
				like(bugReports.hostname, `%${search}%`),
				like(bugReports.os_user, `%${search}%`),
				like(bugReports.name, `%${search}%`),
				like(bugReports.email, `%${search}%`)
			)
		);
	}

	const where = conditions.length > 0 ? and(...conditions) : undefined;

	// Get total count
	const [{ total }] = await db
		.select({ total: count() })
		.from(bugReports)
		.where(where);

	// Get paginated reports with file count
	const reports = await db
		.select({
			id: bugReports.id,
			name: bugReports.name,
			email: bugReports.email,
			hostname: bugReports.hostname,
			os_user: bugReports.os_user,
			status: bugReports.status,
			created_at: bugReports.created_at,
			file_count: count(bugReportFiles.id)
		})
		.from(bugReports)
		.leftJoin(bugReportFiles, eq(bugReports.id, bugReportFiles.report_id))
		.where(where)
		.groupBy(bugReports.id)
		.orderBy(desc(bugReports.created_at))
		.limit(pageSize)
		.offset((page - 1) * pageSize);

	return {
		reports,
		pagination: {
			page,
			pageSize,
			total,
			totalPages: Math.ceil(total / pageSize)
		},
		filters: {
			status,
			search
		}
	};
};
