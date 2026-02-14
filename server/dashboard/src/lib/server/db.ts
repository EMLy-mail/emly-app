import { drizzle } from 'drizzle-orm/mysql2';
import mysql from 'mysql2/promise';
import * as schema from '$lib/schema';
import { env } from '$env/dynamic/private';

const pool = mysql.createPool({
	host: env.MYSQL_HOST || 'localhost',
	port: Number(env.MYSQL_PORT) || 3306,
	user: env.MYSQL_USER || 'emly',
	password: env.MYSQL_PASSWORD,
	database: env.MYSQL_DATABASE || 'emly_bugreports',
	connectionLimit: 10,
	idleTimeout: 60000
});

export const db = drizzle(pool, { schema, mode: 'default' });
