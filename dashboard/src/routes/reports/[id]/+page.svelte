<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { statusColors, statusLabels, formatDate, formatBytes } from '$lib/utils';
	import {
		ArrowLeft,
		Download,
		Trash2,
		FileText,
		Image,
		Monitor,
		Settings,
		Database,
		ChevronDown,
		ChevronRight,
		Mail
	} from 'lucide-svelte';

	let { data } = $props();
	let showDeleteDialog = $state(false);
	let showSystemInfo = $state(false);
	let statusUpdating = $state(false);
	let deleting = $state(false);

	const roleIcons: Record<string, typeof FileText> = {
		screenshot: Image,
		mail_file: Mail,
		localstorage: Database,
		config: Settings,
		system_info: Monitor
	};

	const roleLabels: Record<string, string> = {
		screenshot: 'Screenshot',
		mail_file: 'Mail File',
		localstorage: 'Local Storage',
		config: 'Config',
		system_info: 'System Info'
	};

	async function updateStatus(newStatus: string) {
		statusUpdating = true;
		try {
			const res = await fetch(`/reports/${data.report.id}`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status: newStatus })
			});
			if (res.ok) await invalidateAll();
		} finally {
			statusUpdating = false;
		}
	}

	async function deleteReport() {
		deleting = true;
		try {
			const res = await fetch(`/reports/${data.report.id}`, { method: 'DELETE' });
			if (res.ok) goto('/');
		} finally {
			deleting = false;
		}
	}
</script>

<div class="space-y-6">
	<!-- Back button -->
	<a
		href="/"
		class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
	>
		<ArrowLeft class="h-4 w-4" />
		Back to Reports
	</a>

	<!-- Header card -->
	<div class="rounded-lg border border-border bg-card p-6">
		<div class="flex items-start justify-between">
			<div>
				<div class="flex items-center gap-3">
					<h2 class="text-xl font-semibold">Report #{data.report.id}</h2>
					<span
						class="inline-flex rounded-full border px-2.5 py-0.5 text-xs font-medium {statusColors[data.report.status]}"
					>
						{statusLabels[data.report.status]}
					</span>
				</div>
				<p class="mt-1 text-sm text-muted-foreground">
					Submitted by {data.report.name} ({data.report.email})
				</p>
			</div>
			<div class="flex items-center gap-2">
				<!-- Status selector -->
				<select
					value={data.report.status}
					onchange={(e) => updateStatus(e.currentTarget.value)}
					disabled={statusUpdating}
					class="rounded-md border border-input bg-background px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring disabled:opacity-50"
				>
					<option value="new">New</option>
					<option value="in_review">In Review</option>
					<option value="resolved">Resolved</option>
					<option value="closed">Closed</option>
				</select>

				<!-- Download ZIP -->
				<a
					href="/api/reports/{data.report.id}/download"
					class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90"
				>
					<Download class="h-4 w-4" />
					ZIP
				</a>

				<!-- Delete -->
				<button
					onclick={() => (showDeleteDialog = true)}
					class="inline-flex items-center gap-1.5 rounded-md border border-destructive px-3 py-1.5 text-sm font-medium text-destructive-foreground bg-destructive hover:bg-destructive/80"
				>
					<Trash2 class="h-4 w-4" />
					Delete
				</button>
			</div>
		</div>

		<!-- Metadata grid -->
		<div class="mt-6 grid grid-cols-2 gap-4 md:grid-cols-4">
			<div>
				<p class="text-xs font-medium text-muted-foreground">Hostname</p>
				<p class="mt-1 text-sm">{data.report.hostname || '—'}</p>
			</div>
			<div>
				<p class="text-xs font-medium text-muted-foreground">OS User</p>
				<p class="mt-1 text-sm">{data.report.os_user || '—'}</p>
			</div>
			<div>
				<p class="text-xs font-medium text-muted-foreground">HWID</p>
				<p class="mt-1 font-mono text-xs break-all">{data.report.hwid || '—'}</p>
			</div>
			<div>
				<p class="text-xs font-medium text-muted-foreground">IP Address</p>
				<p class="mt-1 text-sm">{data.report.submitter_ip || '—'}</p>
			</div>
			<div>
				<p class="text-xs font-medium text-muted-foreground">Created</p>
				<p class="mt-1 text-sm">{formatDate(data.report.created_at)}</p>
			</div>
			<div>
				<p class="text-xs font-medium text-muted-foreground">Updated</p>
				<p class="mt-1 text-sm">{formatDate(data.report.updated_at)}</p>
			</div>
		</div>
	</div>

	<!-- Description -->
	<div class="rounded-lg border border-border bg-card p-6">
		<h3 class="text-sm font-medium text-muted-foreground">Description</h3>
		<p class="mt-2 whitespace-pre-wrap text-sm">{data.report.description}</p>
	</div>

	<!-- System Info -->
	{#if data.report.system_info}
		<div class="rounded-lg border border-border bg-card">
			<button
				onclick={() => (showSystemInfo = !showSystemInfo)}
				class="flex w-full items-center gap-2 px-6 py-4 text-left text-sm font-medium text-muted-foreground hover:text-foreground"
			>
				{#if showSystemInfo}
					<ChevronDown class="h-4 w-4" />
				{:else}
					<ChevronRight class="h-4 w-4" />
				{/if}
				System Information
			</button>
			{#if showSystemInfo}
				<div class="border-t border-border px-6 py-4">
					<pre class="overflow-auto rounded-md bg-muted/50 p-4 text-xs">{data.report.system_info}</pre>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Files -->
	<div class="rounded-lg border border-border bg-card p-6">
		<h3 class="mb-4 text-sm font-medium text-muted-foreground">
			Attached Files ({data.files.length})
		</h3>
		{#if data.files.length > 0}
			<!-- Screenshot previews -->
			{@const screenshots = data.files.filter((f) => f.file_role === 'screenshot')}
			{#if screenshots.length > 0}
				<div class="mb-4 grid grid-cols-1 gap-4 md:grid-cols-2">
					{#each screenshots as file}
						<div class="overflow-hidden rounded-md border border-border">
							<img
								src="/api/reports/{data.report.id}/files/{file.id}"
								alt={file.filename}
								class="w-full"
							/>
							<div class="px-3 py-2 text-xs text-muted-foreground">{file.filename}</div>
						</div>
					{/each}
				</div>
			{/if}

			<div class="overflow-hidden rounded-md border border-border">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-border bg-muted/50">
							<th class="px-4 py-2 text-left text-xs font-medium text-muted-foreground">Role</th>
							<th class="px-4 py-2 text-left text-xs font-medium text-muted-foreground">Filename</th>
							<th class="px-4 py-2 text-left text-xs font-medium text-muted-foreground">Size</th>
							<th class="px-4 py-2 text-right text-xs font-medium text-muted-foreground">Action</th>
						</tr>
					</thead>
					<tbody>
						{#each data.files as file}
							{@const Icon = roleIcons[file.file_role] || FileText}
							<tr class="border-b border-border last:border-0">
								<td class="px-4 py-2">
									<span class="inline-flex items-center gap-1.5 text-muted-foreground">
										<Icon class="h-3.5 w-3.5" />
										{roleLabels[file.file_role] || file.file_role}
									</span>
								</td>
								<td class="px-4 py-2 font-mono text-xs">{file.filename}</td>
								<td class="px-4 py-2 text-muted-foreground">{formatBytes(file.file_size)}</td>
								<td class="px-4 py-2 text-right">
									<a
										href="/api/reports/{data.report.id}/files/{file.id}"
										class="inline-flex items-center gap-1 rounded-md border border-input px-2 py-1 text-xs hover:bg-accent"
									>
										<Download class="h-3 w-3" />
										Download
									</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else}
			<p class="text-sm text-muted-foreground">No files attached.</p>
		{/if}
	</div>
</div>

<!-- Delete confirmation dialog -->
{#if showDeleteDialog}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
		onkeydown={(e) => e.key === 'Escape' && (showDeleteDialog = false)}
	>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div class="absolute inset-0" onclick={() => (showDeleteDialog = false)}></div>
		<div class="relative rounded-lg border border-border bg-card p-6 shadow-xl max-w-md w-full">
			<h3 class="text-lg font-semibold">Delete Report</h3>
			<p class="mt-2 text-sm text-muted-foreground">
				Are you sure you want to delete report #{data.report.id}? This will permanently remove the
				report and all attached files. This action cannot be undone.
			</p>
			<div class="mt-4 flex justify-end gap-2">
				<button
					onclick={() => (showDeleteDialog = false)}
					class="rounded-md border border-input px-4 py-2 text-sm hover:bg-accent"
				>
					Cancel
				</button>
				<button
					onclick={deleteReport}
					disabled={deleting}
					class="rounded-md bg-destructive px-4 py-2 text-sm font-medium text-destructive-foreground hover:bg-destructive/80 disabled:opacity-50"
				>
					{deleting ? 'Deleting...' : 'Delete'}
				</button>
			</div>
		</div>
	</div>
{/if}
