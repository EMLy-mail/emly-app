import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { db } from '$lib/server/db';
import { bugReportFiles } from '$lib/schema';
import { eq, and } from 'drizzle-orm';

export const GET: RequestHandler = async ({ params }) => {
	const reportId = Number(params.id);
	const fileId = Number(params.fileId);

	if (isNaN(reportId) || isNaN(fileId)) throw error(400, 'Invalid ID');

	const [file] = await db
		.select()
		.from(bugReportFiles)
		.where(and(eq(bugReportFiles.id, fileId), eq(bugReportFiles.report_id, reportId)))
		.limit(1);

	if (!file) throw error(404, 'File not found');

	return new Response(new Uint8Array(file.data), {
		headers: {
			'Content-Type': file.mime_type,
			'Content-Disposition': `inline; filename="${file.filename}"`,
			'Content-Length': String(file.file_size)
		}
	});
};
