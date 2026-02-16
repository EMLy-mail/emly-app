import type { PageServerLoad } from './$types';
import { db } from '$lib/server/db';
import { bugReports, bugReportFiles } from '$lib/schema';
import { eq } from 'drizzle-orm';
import { error } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ params, locals }) => {
	const id = Number(params.id);
	if (isNaN(id)) throw error(400, 'Invalid report ID');

	const [report] = await db
		.select()
		.from(bugReports)
		.where(eq(bugReports.id, id))
		.limit(1);

	if (!report) throw error(404, 'Report not found');

	const files = await db
		.select({
			id: bugReportFiles.id,
			report_id: bugReportFiles.report_id,
			file_role: bugReportFiles.file_role,
			filename: bugReportFiles.filename,
			mime_type: bugReportFiles.mime_type,
			file_size: bugReportFiles.file_size,
			created_at: bugReportFiles.created_at
		})
		.from(bugReportFiles)
		.where(eq(bugReportFiles.report_id, id));

	return {
		report: {
			...report,
			system_info: report.system_info ? JSON.stringify(report.system_info, null, 2) : null,
			created_at: report.created_at.toISOString(),
			updated_at: report.updated_at.toISOString()
		},
		files: files.map((f) => ({
			...f,
			created_at: f.created_at.toISOString()
		})),
		currentUserId: locals.user?.id ?? ''
	};
};
