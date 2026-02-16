import {
	mysqlTable,
	int,
	varchar,
	text,
	json,
	mysqlEnum,
	timestamp,
	datetime,
	boolean,
	customType
} from 'drizzle-orm/mysql-core';

const longblob = customType<{ data: Buffer }>({
	dataType() {
		return 'longblob';
	}
});

export const bugReports = mysqlTable('bug_reports', {
	id: int('id').autoincrement().primaryKey(),
	name: varchar('name', { length: 255 }).notNull(),
	email: varchar('email', { length: 255 }).notNull(),
	description: text('description').notNull(),
	hwid: varchar('hwid', { length: 255 }).notNull().default(''),
	hostname: varchar('hostname', { length: 255 }).notNull().default(''),
	os_user: varchar('os_user', { length: 255 }).notNull().default(''),
	submitter_ip: varchar('submitter_ip', { length: 45 }).notNull().default(''),
	system_info: json('system_info'),
	status: mysqlEnum('status', ['new', 'in_review', 'resolved', 'closed']).notNull().default('new'),
	created_at: timestamp('created_at').notNull().defaultNow(),
	updated_at: timestamp('updated_at').notNull().defaultNow().onUpdateNow()
});

export const bugReportFiles = mysqlTable('bug_report_files', {
	id: int('id').autoincrement().primaryKey(),
	report_id: int('report_id')
		.notNull()
		.references(() => bugReports.id, { onDelete: 'cascade' }),
	file_role: mysqlEnum('file_role', [
		'screenshot',
		'mail_file',
		'localstorage',
		'config',
		'system_info'
	]).notNull(),
	filename: varchar('filename', { length: 255 }).notNull(),
	mime_type: varchar('mime_type', { length: 127 }).notNull().default('application/octet-stream'),
	file_size: int('file_size').notNull().default(0),
	data: longblob('data').notNull(),
	created_at: timestamp('created_at').notNull().defaultNow()
});

export const userTable = mysqlTable('user', {
	id: varchar('id', { length: 255 }).primaryKey(),
	username: varchar('username', { length: 255 }).notNull().unique(),
	displayname: varchar('displayname', { length: 255 }).notNull().default(''),
	passwordHash: varchar('password_hash', { length: 255 }).notNull(),
	role: mysqlEnum('role', ['admin', 'user']).notNull().default('user'),
	enabled: boolean('enabled').notNull().default(true),
	createdAt: timestamp('created_at').notNull().defaultNow()
});

export const sessionTable = mysqlTable('session', {
	id: varchar('id', { length: 255 }).primaryKey(),
	userId: varchar('user_id', { length: 255 })
		.notNull()
		.references(() => userTable.id),
	expiresAt: datetime('expires_at').notNull()
});

export type BugReport = typeof bugReports.$inferSelect;
export type BugReportFile = typeof bugReportFiles.$inferSelect;
export type BugReportStatus = BugReport['status'];
