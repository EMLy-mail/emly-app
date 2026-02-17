import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { hash } from '@node-rs/argon2';
import { generateIdFromEntropySize } from 'lucia';
import { db } from '$lib/server/db';
import { userTable } from '$lib/schema';
import { eq } from 'drizzle-orm';

const PASSWORD_MIN_LENGTH = 8;
const PASSWORD_MAX_LENGTH = 255;

function validatePassword(password: string): string | null {
	if (password.length < PASSWORD_MIN_LENGTH || password.length > PASSWORD_MAX_LENGTH) {
		return `Password must be ${PASSWORD_MIN_LENGTH}-${PASSWORD_MAX_LENGTH} characters`;
	}
	if (!/[A-Z]/.test(password)) {
		return 'Password must contain at least one uppercase letter';
	}
	if (!/[a-z]/.test(password)) {
		return 'Password must contain at least one lowercase letter';
	}
	if (!/[0-9]/.test(password)) {
		return 'Password must contain at least one number';
	}
	if (!/[^A-Za-z0-9]/.test(password)) {
		return 'Password must contain at least one special character';
	}
	return null;
}

async function hashPassword(password: string): Promise<string> {
	return hash(password, {
		memoryCost: 19456,
		timeCost: 2,
		outputLen: 32,
		parallelism: 1
	});
}

export const load: PageServerLoad = async ({ locals }) => {
	if (!locals.user || locals.user.role !== 'admin') {
		redirect(302, '/');
	}

	const users = await db
		.select({
			id: userTable.id,
			username: userTable.username,
			displayname: userTable.displayname,
			role: userTable.role,
			enabled: userTable.enabled,
			createdAt: userTable.createdAt
		})
		.from(userTable)
		.orderBy(userTable.createdAt);

	return { users };
};

export const actions: Actions = {
	create: async ({ request, locals }) => {
		if (!locals.user || locals.user.role !== 'admin') {
			return fail(403, { message: 'Unauthorized' });
		}

		const formData = await request.formData();
		const username = formData.get('username');
		const displayname = formData.get('displayname') || '';
		const password = formData.get('password');
		const confirmPassword = formData.get('confirmPassword');
		const role = formData.get('role');

		if (
			typeof username !== 'string' ||
			typeof displayname !== 'string' ||
			typeof password !== 'string' ||
			typeof confirmPassword !== 'string' ||
			typeof role !== 'string'
		) {
			return fail(400, { message: 'Invalid input' });
		}

		if (!username || !password) {
			return fail(400, { message: 'Username and password are required' });
		}

		if (username.length < 3 || username.length > 255) {
			return fail(400, { message: 'Username must be 3-255 characters' });
		}

		if (password !== confirmPassword) {
			return fail(400, { message: 'Passwords do not match' });
		}

		const passwordError = validatePassword(password);
		if (passwordError) {
			return fail(400, { message: passwordError });
		}

		if (role !== 'admin' && role !== 'user') {
			return fail(400, { message: 'Invalid role' });
		}

		// Check if username already exists
		const [existing] = await db
			.select({ id: userTable.id })
			.from(userTable)
			.where(eq(userTable.username, username))
			.limit(1);

		if (existing) {
			return fail(400, { message: 'Username already exists' });
		}

		const passwordHash = await hashPassword(password);
		const userId = generateIdFromEntropySize(10);

		await db.insert(userTable).values({
			id: userId,
			username,
			displayname,
			passwordHash,
			role: role as 'admin' | 'user'
		});

		return { success: true };
	},

	updateDisplayname: async ({ request, locals }) => {
		if (!locals.user || locals.user.role !== 'admin') {
			return fail(403, { message: 'Unauthorized' });
		}

		const formData = await request.formData();
		const userId = formData.get('userId');
		const displayname = formData.get('displayname');

		if (typeof userId !== 'string' || typeof displayname !== 'string') {
			return fail(400, { message: 'Invalid input' });
		}

		await db.update(userTable).set({ displayname }).where(eq(userTable.id, userId));

		return { success: true };
	},

	resetPassword: async ({ request, locals }) => {
		if (!locals.user || locals.user.role !== 'admin') {
			return fail(403, { message: 'Unauthorized' });
		}

		const formData = await request.formData();
		const userId = formData.get('userId');

		if (typeof userId === 'string' && userId === locals.user.id) {
			return fail(400, { message: 'Cannot reset your own password from here' });
		}
		const newPassword = formData.get('newPassword');
		const confirmPassword = formData.get('confirmPassword');

		if (
			typeof userId !== 'string' ||
			typeof newPassword !== 'string' ||
			typeof confirmPassword !== 'string'
		) {
			return fail(400, { message: 'Invalid input' });
		}

		if (newPassword !== confirmPassword) {
			return fail(400, { message: 'Passwords do not match' });
		}

		const passwordError = validatePassword(newPassword);
		if (passwordError) {
			return fail(400, { message: passwordError });
		}

		const passwordHash = await hashPassword(newPassword);

		await db.update(userTable).set({ passwordHash }).where(eq(userTable.id, userId));

		return { success: true };
	},

	toggleEnabled: async ({ request, locals }) => {
		if (!locals.user || locals.user.role !== 'admin') {
			return fail(403, { message: 'Unauthorized' });
		}

		const formData = await request.formData();
		const userId = formData.get('userId');

		if (typeof userId !== 'string') {
			return fail(400, { message: 'Invalid input' });
		}

		// Cannot disable yourself
		if (userId === locals.user.id) {
			return fail(400, { message: 'Cannot disable your own account' });
		}

		// Cannot disable other admins
		const [targetUser] = await db
			.select({ role: userTable.role, enabled: userTable.enabled })
			.from(userTable)
			.where(eq(userTable.id, userId))
			.limit(1);

		if (!targetUser) {
			return fail(404, { message: 'User not found' });
		}

		if (targetUser.role === 'admin') {
			return fail(400, { message: 'Cannot disable an admin user' });
		}

		await db
			.update(userTable)
			.set({ enabled: !targetUser.enabled })
			.where(eq(userTable.id, userId));

		return { success: true };
	},

	delete: async ({ request, locals }) => {
		if (!locals.user || locals.user.role !== 'admin') {
			return fail(403, { message: 'Unauthorized' });
		}

		const formData = await request.formData();
		const userId = formData.get('userId');

		if (typeof userId !== 'string') {
			return fail(400, { message: 'Invalid input' });
		}

		if (userId === locals.user.id) {
			return fail(400, { message: 'Cannot delete your own account' });
		}

		// Prevent deleting admin users
		const [targetUser] = await db
			.select({ role: userTable.role })
			.from(userTable)
			.where(eq(userTable.id, userId))
			.limit(1);

		if (targetUser?.role === 'admin') {
			return fail(400, { message: 'Cannot delete an admin user' });
		}

		await db.delete(userTable).where(eq(userTable.id, userId));

		return { success: true };
	}
};
