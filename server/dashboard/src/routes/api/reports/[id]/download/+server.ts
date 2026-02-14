import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { db } from '$lib/server/db';
import { bugReports, bugReportFiles } from '$lib/schema';
import { eq } from 'drizzle-orm';
import JSZip from 'jszip';

export const GET: RequestHandler = async ({ params }) => {
	const id = Number(params.id);
	if (isNaN(id)) throw error(400, 'Invalid report ID');

	const [report] = await db
		.select()
		.from(bugReports)
		.where(eq(bugReports.id, id))
		.limit(1);

	if (!report) throw error(404, 'Report not found');

	const files = await db
		.select()
		.from(bugReportFiles)
		.where(eq(bugReportFiles.report_id, id));

	const zip = new JSZip();

	// Add report metadata as text file
	const reportText = [
		`Bug Report #${report.id}`,
		`========================`,
		``,
		`Name: ${report.name}`,
		`Email: ${report.email}`,
		`Hostname: ${report.hostname}`,
		`OS User: ${report.os_user}`,
		`HWID: ${report.hwid}`,
		`IP: ${report.submitter_ip}`,
		`Status: ${report.status}`,
		`Created: ${report.created_at.toISOString()}`,
		`Updated: ${report.updated_at.toISOString()}`,
		``,
		`Description:`,
		`------------`,
		report.description,
		``,
		...(report.system_info
			? [`System Info:`, `------------`, JSON.stringify(report.system_info, null, 2)]
			: [])
	].join('\n');

	zip.file('report.txt', reportText);

	// Add all files
	for (const file of files) {
		const folder = file.file_role;
		zip.file(`${folder}/${file.filename}`, file.data);
	}

	const zipBuffer = await zip.generateAsync({ type: 'nodebuffer' });

	return new Response(new Uint8Array(zipBuffer), {
		headers: {
			'Content-Type': 'application/zip',
			'Content-Disposition': `attachment; filename="report-${id}.zip"`,
			'Content-Length': String(zipBuffer.length)
		}
	});
};
