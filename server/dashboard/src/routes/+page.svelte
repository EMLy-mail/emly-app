<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/stores';
	import { statusColors, statusLabels, formatDate } from '$lib/utils';
	import { Search, ChevronLeft, ChevronRight, Filter, Paperclip, RefreshCcw, Inbox, Upload } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Table from '$lib/components/ui/table';
	import * as Select from '$lib/components/ui/select';
	import * as Empty from '$lib/components/ui/empty';
	import * as Dialog from '$lib/components/ui/dialog';

	let { data } = $props();

	let searchInput = $state('');
	let statusFilter = $state('');

	const statusOptions = [
		{ value: 'new', label: 'New' },
		{ value: 'in_review', label: 'In Review' },
		{ value: 'resolved', label: 'Resolved' },
		{ value: 'closed', label: 'Closed' }
	];

	$effect(() => {
		searchInput = data.filters.search;
		statusFilter = data.filters.status;
	});

	function applyFilters() {
		const params = new URLSearchParams();
		if (searchInput) params.set('search', searchInput);
		if (statusFilter) params.set('status', statusFilter);
		params.set('page', '1');
		goto(`/?${params.toString()}`);
	}

	function goToPage(p: number) {
		const params = new URLSearchParams($page.url.searchParams);
		params.set('page', String(p));
		goto(`/?${params.toString()}`);
	}

	function clearFilters() {
		searchInput = '';
		statusFilter = '';
		goto('/');
	}

	async function refreshReports() {
		try {
			await invalidateAll();
		} catch (err) {
			console.error('Failed to refresh reports:', err);
		}
	}
</script>

<div class="space-y-4">
	<!-- Filters -->
	<div class="flex flex-wrap items-center gap-3">
		<div class="relative flex-1 min-w-50 max-w-sm">
			<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="text"
				placeholder="Search hostname, user, name, email..."
				bind:value={searchInput}
				onkeydown={(e) => e.key === 'Enter' && applyFilters()}
				class="w-full rounded-md border border-input bg-background py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
			/>
		</div>
		<Select.Root type="single" bind:value={statusFilter} onValueChange={() => applyFilters()}>
			<Select.Trigger class="w-37.5">
				{statusOptions.find((o) => o.value === statusFilter)?.label || 'All statuses'}
			</Select.Trigger>
			<Select.Content>
				<Select.Item value="" label="All statuses" />
				{#each statusOptions as option}
					<Select.Item value={option.value} label={option.label} />
				{/each}
			</Select.Content>
		</Select.Root>
		<Button onclick={applyFilters}>
			<Filter class="h-4 w-4" />
			Filter
		</Button>
		<Button onclick={refreshReports}>
			<RefreshCcw class="h-4 w-4" />
			Refresh
		</Button>
		{#if data.filters.search || data.filters.status}
			<Button variant="outline" onclick={clearFilters}>
				Clear
			</Button>
		{/if}
	</div>

	<!-- Table -->
	{#if data.reports.length === 0}
		<div class="rounded-lg border border-border bg-card">
			<Empty.Root class="p-8">
				<Empty.Header>
					<Empty.Media variant="icon">
						<Inbox class="h-10 w-10 text-muted-foreground" />
					</Empty.Media>
					<Empty.Title>No reports found</Empty.Title>
					<Empty.Description>
						There are no bug reports matching your current filters.
					</Empty.Description>
				</Empty.Header>
				{#if data.filters.search || data.filters.status}
					<Empty.Content>
						<Button variant="outline" onclick={clearFilters}>
							Clear Filters
						</Button>
					</Empty.Content>
				{:else}
					<Empty.Content>
						<Button variant="outline" onclick={refreshReports}>
							<RefreshCcw class="mr-2 h-4 w-4" />
							Refresh
						</Button>
					</Empty.Content>
				{/if}
			</Empty.Root>
		</div>
	{:else}
		<div class="overflow-hidden rounded-lg border border-border">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head class="w-[80px]">ID</Table.Head>
						<Table.Head>Hostname</Table.Head>
						<Table.Head>User</Table.Head>
						<Table.Head>Reporter</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head>Files</Table.Head>
						<Table.Head class="text-right">Created</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.reports as report (report.id)}
						<Table.Row
							class="cursor-pointer"
							onclick={() => goto(`/reports/${report.id}`)}
						>
							<Table.Cell class="font-mono text-muted-foreground">#{report.id}</Table.Cell>
							<Table.Cell>{report.hostname || '—'}</Table.Cell>
							<Table.Cell>{report.os_user || '—'}</Table.Cell>
							<Table.Cell>
								<div class="font-medium">{report.name}</div>
								<div class="text-xs text-muted-foreground">{report.email}</div>
							</Table.Cell>
							<Table.Cell>
								<span
									class="inline-flex rounded-full border px-2 py-0.5 text-xs font-medium {statusColors[
										report.status
									]}"
								>
									{statusLabels[report.status]}
								</span>
							</Table.Cell>
							<Table.Cell>
								{#if report.file_count > 0}
									<span class="inline-flex items-center gap-1 text-muted-foreground">
										<Paperclip class="h-3.5 w-3.5" />
										{report.file_count}
									</span>
								{:else}
									<span class="text-muted-foreground">—</span>
								{/if}
							</Table.Cell>
							<Table.Cell class="text-right text-muted-foreground">
								{formatDate(report.created_at)}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	{/if}

	<!-- Pagination -->
	{#if data.pagination.totalPages > 1}
		<div class="flex items-center justify-between">
			<p class="text-sm text-muted-foreground">
				Showing {(data.pagination.page - 1) * data.pagination.pageSize + 1} to {Math.min(
					data.pagination.page * data.pagination.pageSize,
					data.pagination.total
				)} of {data.pagination.total} reports
			</p>
			<div class="flex items-center gap-1">
				<Button
					variant="outline"
					size="icon"
					onclick={() => goToPage(data.pagination.page - 1)}
					disabled={data.pagination.page <= 1}
				>
					<ChevronLeft class="h-4 w-4" />
				</Button>
				{#each Array.from({ length: data.pagination.totalPages }, (_, i) => i + 1) as p}
					{#if p === 1 || p === data.pagination.totalPages || (p >= data.pagination.page - 1 && p <= data.pagination.page + 1)}
						<Button
							variant={p === data.pagination.page ? 'default' : 'outline'}
							size="icon"
							onclick={() => goToPage(p)}
						>
							{p}
						</Button>
					{:else if p === data.pagination.page - 2 || p === data.pagination.page + 2}
						<span class="px-1 text-muted-foreground">...</span>
					{/if}
				{/each}
				<Button
					variant="outline"
					size="icon"
					onclick={() => goToPage(data.pagination.page + 1)}
					disabled={data.pagination.page >= data.pagination.totalPages}
				>
					<ChevronRight class="h-4 w-4" />
				</Button>
			</div>
		</div>
	{/if}
</div>

