<script lang="ts">
	import { enhance } from '$app/forms';
	import { Bug } from 'lucide-svelte';

	let { form } = $props();
</script>

<div class="flex min-h-screen items-center justify-center bg-background">
	<div class="w-full max-w-sm rounded-lg border border-border bg-card p-8">
		<div class="mb-6 flex flex-col items-center gap-2">
			<Bug class="h-10 w-10 text-primary" />
			<h1 class="text-xl font-semibold text-foreground">EMLy Dashboard</h1>
			<p class="text-sm text-muted-foreground">Sign in to continue</p>
		</div>

		{#if form?.message}
			<div class="mb-4 rounded-md bg-destructive/10 px-3 py-2 text-sm text-red-400">
				{form.message}
			</div>
		{/if}

		<form method="POST" use:enhance class="space-y-4">
			<div>
				<label for="username" class="mb-1 block text-sm font-medium text-foreground">
					Username
				</label>
				<input
					type="text"
					id="username"
					name="username"
					required
					autocomplete="username"
					class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					placeholder="Enter username"
				/>
			</div>

			<div>
				<label for="password" class="mb-1 block text-sm font-medium text-foreground">
					Password
				</label>
				<input
					type="password"
					id="password"
					name="password"
					required
					autocomplete="current-password"
					class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
					placeholder="Enter password"
				/>
			</div>

			<button
				type="submit"
				class="w-full rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
			>
				Sign in
			</button>
		</form>
	</div>
</div>
