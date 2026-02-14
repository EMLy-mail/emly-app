import type { RequestHandler } from './$types';
import { json } from '@sveltejs/kit';
import { db } from '$lib/server/db';
import { bugReports } from '$lib/schema';
import { count } from 'drizzle-orm';

export const GET: RequestHandler = async () => {
    const [{ total }] = await db
        .select({ total: count() })
        .from(bugReports);

    return json({ total });
};