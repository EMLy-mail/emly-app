<script lang="ts">
	import { enhance } from '$app/forms';
	import { formatDate } from '$lib/utils';
	import { Trash2, UserPlus, Pencil, KeyRound, Check, X, ShieldOff, ShieldCheck } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';

	let { data, form } = $props();

	// Create user dialog visibility
	let showCreateForm = $state(false);

	// Password policy state
	let password = $state('');
	let confirmPassword = $state('');
	let role = $state('user');

	const policies = $derived({
		minLength: password.length >= 8,
		uppercase: /[A-Z]/.test(password),
		lowercase: /[a-z]/.test(password),
		number: /[0-9]/.test(password),
		special: /[^A-Za-z0-9]/.test(password),
		match: password.length > 0 && password === confirmPassword
	});

	const allPoliciesMet = $derived(
		policies.minLength &&
			policies.uppercase &&
			policies.lowercase &&
			policies.number &&
			policies.special &&
			policies.match
	);

	// Dialog states
	let displaynameDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
	let selectedUser = $state<{ id: string; username: string; displayname: string } | null>(null);
	let newDisplayname = $state('');

	function openDisplaynameDialog(user: typeof selectedUser) {
		selectedUser = user;
		newDisplayname = user?.displayname || '';
		displaynameDialogOpen = true;
	}

	function openDeleteDialog(user: typeof selectedUser) {
		selectedUser = user;
		deleteDialogOpen = true;
	}

	// Reset password dialog state
	let resetDialogOpen = $state(false);
	let resetUserId = $state('');
	let resetUsername = $state('');
	let resetPassword = $state('');
	let resetConfirmPassword = $state('');

	const resetPolicies = $derived({
		minLength: resetPassword.length >= 8,
		uppercase: /[A-Z]/.test(resetPassword),
		lowercase: /[a-z]/.test(resetPassword),
		number: /[0-9]/.test(resetPassword),
		special: /[^A-Za-z0-9]/.test(resetPassword),
		match: resetPassword.length > 0 && resetPassword === resetConfirmPassword
	});

	const allResetPoliciesMet = $derived(
		resetPolicies.minLength &&
			resetPolicies.uppercase &&
			resetPolicies.lowercase &&
			resetPolicies.number &&
			resetPolicies.special &&
			resetPolicies.match
	);

	function openResetDialog(userId: string, username: string) {
		resetUserId = userId;
		resetUsername = username;
		resetPassword = '';
		resetConfirmPassword = '';
		resetDialogOpen = true;
	}
</script>

<div class="space-y-6">
	<Button onclick={() => (showCreateForm = true)}>
		<UserPlus class="h-4 w-4" />
		Create User
	</Button>

	<!-- Users Table -->
	<div class="overflow-hidden rounded-lg border border-border">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-border bg-muted/50">
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Username</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Display Name</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Role</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Created</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each data.users as user (user.id)}
					<tr class="border-b border-border transition-colors hover:bg-muted/30">
						<td class="px-4 py-3 font-mono">{user.username}</td>
						<td class="px-4 py-3">{user.displayname || '—'}</td>
						<td class="px-4 py-3">
							<span
								class="inline-flex rounded-full border px-2 py-0.5 text-xs font-medium {user.role ===
								'admin'
									? 'bg-purple-500/20 text-purple-400 border-purple-500/30'
									: 'bg-zinc-500/20 text-zinc-400 border-zinc-500/30'}"
							>
								{user.role}
							</span>
						</td>
						<td class="px-4 py-3">
							<span
								class="inline-flex rounded-full border px-2 py-0.5 text-xs font-medium {user.enabled
									? 'bg-green-500/20 text-green-400 border-green-500/30'
									: 'bg-red-500/20 text-red-400 border-red-500/30'}"
							>
								{user.enabled ? 'Enabled' : 'Disabled'}
							</span>
						</td>
						<td class="px-4 py-3 text-muted-foreground">
							{user.createdAt ? formatDate(user.createdAt) : '—'}
						</td>
						<td class="px-4 py-3">
							{#if user.id === data.user?.id}
								<span class="text-xs text-muted-foreground">Current user</span>
							{:else}
								<div class="flex items-center gap-1">
									{#if user.role !== 'admin'}
										<form method="POST" action="?/toggleEnabled" use:enhance>
											<input type="hidden" name="userId" value={user.id} />
											<Button
												type="submit"
												variant="ghost"
												size="icon"
												class="h-8 w-8 {user.enabled
													? 'text-orange-400 hover:text-orange-500 hover:bg-orange-500/10'
													: 'text-green-400 hover:text-green-500 hover:bg-green-500/10'}"
												title={user.enabled ? 'Disable User' : 'Enable User'}
											>
												{#if user.enabled}
													<ShieldOff class="h-4 w-4" />
												{:else}
													<ShieldCheck class="h-4 w-4" />
												{/if}
											</Button>
										</form>
									{/if}
									<Button
										variant="ghost"
										size="icon"
										class="h-8 w-8"
										onclick={() => openResetDialog(user.id, user.username)}
										title="Change Password"
									>
										<KeyRound class="h-4 w-4" />
									</Button>
									<Button
										variant="ghost"
										size="icon"
										class="h-8 w-8"
										onclick={() => openDisplaynameDialog(user)}
										title="Change Display Name"
									>
										<Pencil class="h-4 w-4" />
									</Button>
									{#if user.id !== data.user?.id && data.user?.role === 'admin'}
										<Button
											variant="ghost"
											size="icon"
											class="h-8 w-8 text-red-400 hover:text-red-500 hover:bg-red-500/10"
											onclick={() => openDeleteDialog(user)}
											title="Delete User"
										>
											<Trash2 class="h-4 w-4" />
										</Button>
									{/if}
								</div>
							{/if}
						</td>
					</tr>
				{:else}
					<tr>
						<td colspan="6" class="px-4 py-12 text-center text-muted-foreground">
							No users found.
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

<!-- Create User Dialog -->
<Dialog.Root bind:open={showCreateForm}>
	<Dialog.Content class="sm:max-w-lg">
		<Dialog.Header>
			<Dialog.Title>Create User</Dialog.Title>
			<Dialog.Description>Add a new user account with role assignment</Dialog.Description>
		</Dialog.Header>

		<form
			method="POST"
			action="?/create"
			use:enhance={() => {
				return async ({ result, update }) => {
					if (result.type === 'success') {
						showCreateForm = false;
						password = '';
						confirmPassword = '';
						role = 'user';
					}
					await update();
				};
			}}
			class="space-y-4"
		>
			{#if form?.message}
				<div class="rounded-md bg-destructive/10 px-3 py-2 text-sm text-red-400">
					{form.message}
				</div>
			{/if}

			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
				<div class="space-y-2">
					<Label for="username">Username</Label>
					<Input
						type="text"
						id="username"
						name="username"
						required
						minlength={3}
						maxlength={255}
						placeholder="Username"
					/>
				</div>
				<div class="space-y-2">
					<Label for="displayname">Display Name</Label>
					<Input
						type="text"
						id="displayname"
						name="displayname"
						maxlength={255}
						placeholder="Display Name (optional)"
					/>
				</div>
				<div class="space-y-2 sm:col-span-2">
					<Label for="role">Role</Label>
					<input type="hidden" name="role" value={role} />
					<Select.Root type="single" bind:value={role}>
						<Select.Trigger class="w-full">
							{role === 'admin' ? 'Admin' : 'User'}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="user" label="User" />
							<Select.Item value="admin" label="Admin" />
						</Select.Content>
					</Select.Root>
				</div>
			</div>

			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input
						type="password"
						id="password"
						name="password"
						required
						minlength={8}
						maxlength={255}
						placeholder="Password"
						bind:value={password}
					/>
				</div>
				<div class="space-y-2">
					<Label for="confirmPassword">Confirm Password</Label>
					<Input
						type="password"
						id="confirmPassword"
						name="confirmPassword"
						required
						minlength={8}
						maxlength={255}
						placeholder="Confirm password"
						bind:value={confirmPassword}
					/>
				</div>
			</div>

			{#if password.length > 0}
				<div class="grid grid-cols-2 gap-x-4 gap-y-1 text-xs sm:grid-cols-3">
					<span class={policies.minLength ? 'text-green-400' : 'text-muted-foreground'}>
						{policies.minLength ? '✓' : '○'} Min 8 characters
					</span>
					<span class={policies.uppercase ? 'text-green-400' : 'text-muted-foreground'}>
						{policies.uppercase ? '✓' : '○'} Uppercase letter
					</span>
					<span class={policies.lowercase ? 'text-green-400' : 'text-muted-foreground'}>
						{policies.lowercase ? '✓' : '○'} Lowercase letter
					</span>
					<span class={policies.number ? 'text-green-400' : 'text-muted-foreground'}>
						{policies.number ? '✓' : '○'} Number
					</span>
					<span class={policies.special ? 'text-green-400' : 'text-muted-foreground'}>
						{policies.special ? '✓' : '○'} Special character
					</span>
					<span class={policies.match ? 'text-green-400' : 'text-muted-foreground'}>
						{policies.match ? '✓' : '○'} Passwords match
					</span>
				</div>
			{/if}

			<Dialog.Footer>
				<Button type="button" variant="outline" onclick={() => (showCreateForm = false)}>
					Cancel
				</Button>
				<Button type="submit" disabled={!allPoliciesMet}>
					<UserPlus class="h-4 w-4" />
					Create User
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- Reset Password Dialog -->
<Dialog.Root bind:open={resetDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Reset Password</Dialog.Title>
			<Dialog.Description>
				Set a new password for <span class="font-medium text-foreground">{resetUsername}</span>
			</Dialog.Description>
		</Dialog.Header>

		<form
			method="POST"
			action="?/resetPassword"
			use:enhance={() => {
				return async ({ result, update }) => {
					if (result.type === 'success') {
						resetDialogOpen = false;
					}
					await update();
				};
			}}
			class="space-y-4"
		>
			<input type="hidden" name="userId" value={resetUserId} />

			<div class="space-y-2">
				<Label for="resetNewPassword">New Password</Label>
				<Input
					type="password"
					id="resetNewPassword"
					name="newPassword"
					required
					minlength={8}
					maxlength={255}
					placeholder="New password"
					bind:value={resetPassword}
				/>
			</div>

			<div class="space-y-2">
				<Label for="resetConfirmPassword">Confirm New Password</Label>
				<Input
					type="password"
					id="resetConfirmPassword"
					name="confirmPassword"
					required
					minlength={8}
					maxlength={255}
					placeholder="Confirm new password"
					bind:value={resetConfirmPassword}
				/>
			</div>

			{#if resetPassword.length > 0}
				<div class="grid grid-cols-2 gap-x-4 gap-y-1 text-xs">
					<span class={resetPolicies.minLength ? 'text-green-400' : 'text-muted-foreground'}>
						{resetPolicies.minLength ? '✓' : '○'} Min 8 characters
					</span>
					<span class={resetPolicies.uppercase ? 'text-green-400' : 'text-muted-foreground'}>
						{resetPolicies.uppercase ? '✓' : '○'} Uppercase letter
					</span>
					<span class={resetPolicies.lowercase ? 'text-green-400' : 'text-muted-foreground'}>
						{resetPolicies.lowercase ? '✓' : '○'} Lowercase letter
					</span>
					<span class={resetPolicies.number ? 'text-green-400' : 'text-muted-foreground'}>
						{resetPolicies.number ? '✓' : '○'} Number
					</span>
					<span class={resetPolicies.special ? 'text-green-400' : 'text-muted-foreground'}>
						{resetPolicies.special ? '✓' : '○'} Special character
					</span>
					<span class={resetPolicies.match ? 'text-green-400' : 'text-muted-foreground'}>
						{resetPolicies.match ? '✓' : '○'} Passwords match
					</span>
				</div>
			{/if}

			<Dialog.Footer>
				<Button type="button" variant="outline" onclick={() => (resetDialogOpen = false)}>
					Cancel
				</Button>
				<Button type="submit" disabled={!allResetPoliciesMet}>
					Reset Password
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- Update Displayname Dialog -->
<Dialog.Root bind:open={displaynameDialogOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Update Display Name</Dialog.Title>
			<Dialog.Description>
				Update display name for <span class="font-medium text-foreground">{selectedUser?.username}</span>
			</Dialog.Description>
		</Dialog.Header>

		<form
			method="POST"
			action="?/updateDisplayname"
			use:enhance={() => {
				return async ({ result, update }) => {
					if (result.type === 'success') {
						displaynameDialogOpen = false;
					}
					await update();
				};
			}}
			class="space-y-4"
		>
			<input type="hidden" name="userId" value={selectedUser?.id} />

			<div class="space-y-2">
				<Label for="updateDisplayname">Display Name</Label>
				<Input
					type="text"
					id="updateDisplayname"
					name="displayname"
					maxlength={255}
					placeholder="Display Name"
					bind:value={newDisplayname}
				/>
			</div>

			<Dialog.Footer>
				<Button type="button" variant="outline" onclick={() => (displaynameDialogOpen = false)}>
					Cancel
				</Button>
				<Button type="submit">
					Update
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>

<!-- Delete User Alert Dialog -->
<AlertDialog.Root bind:open={deleteDialogOpen}>
	<AlertDialog.Trigger />
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
			<AlertDialog.Description>
				This action cannot be undone. This will permanently delete the user account
				<span class="font-medium text-foreground">{selectedUser?.username}</span> and remove their
				data from our servers.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel onclick={() => (deleteDialogOpen = false)}>Cancel</AlertDialog.Cancel>
			<form method="POST" action="?/delete" use:enhance={() => {
				return async ({ result, update }) => {
					if (result.type === 'success') {
						deleteDialogOpen = false;
					}
					await update();
				};
			}}>
				<input type="hidden" name="userId" value={selectedUser?.id} />
				<Button type="submit" variant="destructive">
					Delete User
				</Button>
			</form>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
