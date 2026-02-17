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
		Mail,
		Eye
	} from 'lucide-svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Card from '$lib/components/ui/card';
	import * as Select from '$lib/components/ui/select';
	import * as Table from '$lib/components/ui/table';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import { presence } from '$lib/stores/presence.svelte';

	let { data } = $props();
	let showDeleteDialog = $state(false);
	let statusUpdating = $state(false);
	let deleting = $state(false);

	const otherViewers = $derived(
		presence.getViewersForReport(data.report.id).filter((u) => u.userId !== data.currentUserId)
	);

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
	<Button variant="ghost" class="gap-1.5 pl-0 hover:bg-transparent hover:text-foreground" href="/">
		<ArrowLeft class="h-4 w-4" />
		Back to Reports
	</Button>

	<!-- Header card -->
	<Card.Root>
		<Card.Header class="pb-4">
			<div class="flex items-start justify-between">
				<div>
					<div class="flex items-center gap-3">
						<Card.Title class="text-xl">Report #{data.report.id}</Card.Title>
						<span
							class="inline-flex rounded-full border px-2.5 py-0.5 text-xs font-medium {statusColors[data.report.status]}"
						>
							{statusLabels[data.report.status]}
						</span>
					</div>
					<Card.Description class="mt-1">
						Submitted by {data.report.name} ({data.report.email})
					</Card.Description>
				</div>
				<div class="flex items-center gap-2">
					{#if otherViewers.length > 0}
						<div class="flex items-center gap-1.5 mr-2">
							<Eye class="h-4 w-4 text-muted-foreground" />
							{#each otherViewers as viewer}
								<Tooltip.Root>
									<Tooltip.Trigger>
										<div class="relative flex h-6 w-6 items-center justify-center rounded-full bg-green-500/20 text-xs font-medium text-green-400 border border-green-500/30">
											{(viewer.displayname || viewer.username).charAt(0).toUpperCase()}
											<span class="absolute -bottom-0.5 -right-0.5 h-2 w-2 rounded-full bg-green-500 border border-background"></span>
										</div>
									</Tooltip.Trigger>
									<Tooltip.Content>
										{viewer.displayname || viewer.username} is viewing this report
									</Tooltip.Content>
								</Tooltip.Root>
							{/each}
						</div>
					{/if}
					<!-- Status selector -->
					<Select.Root
						type="single"
						value={data.report.status}
						disabled={statusUpdating}
						onValueChange={(val) => updateStatus(val)}
					>
						<Select.Trigger class="w-35 disabled:cursor-not-allowed disabled:opacity-50" disabled={statusUpdating}>
							{statusLabels[data.report.status]}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="new" label="New" />
							<Select.Item value="in_review" label="In Review" />
							<Select.Item value="resolved" label="Resolved" />
							<Select.Item value="closed" label="Closed" />
						</Select.Content>
					</Select.Root>

					<!-- Download ZIP -->
					<Button variant="default" class="gap-1.5" href="/api/reports/{data.report.id}/download">
						<Download class="h-4 w-4" />
						ZIP
					</Button>

					<!-- Delete -->
					<Button
						variant="destructive"
						class="gap-1.5"
						onclick={() => (showDeleteDialog = true)}
					>
						<Trash2 class="h-4 w-4" />
						Delete
					</Button>
				</div>
			</div>
		</Card.Header>
		<Card.Content>
			<!-- Metadata grid -->
			<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
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
		</Card.Content>
	</Card.Root>

	<!-- Description -->
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-lg">Description</Card.Title>
		</Card.Header>
		<Card.Content>
			<p class="whitespace-pre-wrap text-sm">{data.report.description}</p>
		</Card.Content>
	</Card.Root>

	<!-- System Info -->
	{#if data.report.system_info}
		<Card.Root class="overflow-hidden">
			<Card.Header>
				<Card.Title class="text-lg">System Information</Card.Title>
			</Card.Header>
			<Card.Content>
				<div>
					<Textarea
						readonly
						class="font-mono text-xs h-128 resize-none"
						value={data.report.system_info}
					/>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}


	<!-- Files -->
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-lg">Attached Files ({data.files.length})</Card.Title>
		</Card.Header>
		<Card.Content>
			{#if data.files.length > 0}
				<!-- Screenshot previews -->
				{@const screenshots = data.files.filter((f) => f.file_role === 'screenshot')}
				{#if screenshots.length > 0}
					<div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-2">
						{#each screenshots as file}
							<div class="overflow-hidden rounded-md border border-border">
								<img
									src="/api/reports/{data.report.id}/files/{file.id}"
									alt={file.filename}
									class="w-full"
								/>
								<div class="bg-muted/50 px-3 py-2 text-xs text-muted-foreground">{file.filename}</div>
							</div>
						{/each}
					</div>
				{/if}

				<div class="rounded-md border">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Role</Table.Head>
								<Table.Head>Filename</Table.Head>
								<Table.Head>Size</Table.Head>
								<Table.Head class="text-right">Action</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each data.files as file}
								{@const Icon = roleIcons[file.file_role] || FileText}
								<Table.Row>
									<Table.Cell>
										<span class="inline-flex items-center gap-2 text-muted-foreground">
											<Icon class="h-4 w-4" />
											{roleLabels[file.file_role] || file.file_role}
										</span>
									</Table.Cell>
									<Table.Cell class="font-mono text-xs">{file.filename}</Table.Cell>
									<Table.Cell class="text-muted-foreground">{formatBytes(file.file_size)}</Table.Cell>
									<Table.Cell class="text-right">
										<Button
											variant="outline"
											size="sm"
											class="h-7 gap-1"
											href="/api/reports/{data.report.id}/files/{file.id}"
										>
											<Download class="h-3 w-3" />
											
										</Button>
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
			{:else}
				<p class="text-sm text-muted-foreground">No files attached.</p>
			{/if}
		</Card.Content>
	</Card.Root>
</div>

<!-- Delete confirmation dialog -->
<AlertDialog.Root bind:open={showDeleteDialog}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete Report</AlertDialog.Title>
			<AlertDialog.Description>
				Are you sure you want to delete report #{data.report.id}? This will permanently remove the
				report and all attached files. This action cannot be undone.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel onclick={() => (showDeleteDialog = false)}>Cancel</AlertDialog.Cancel>
			<Button
				variant="destructive"
				onclick={deleteReport}
				disabled={deleting}
			>
				{deleting ? 'Deleting...' : 'Delete'}
			</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
