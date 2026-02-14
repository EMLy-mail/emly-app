<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/stores';
	import { statusColors, statusLabels, formatDate } from '$lib/utils';
	import { Search, ChevronLeft, ChevronRight, Filter, Paperclip, RefreshCcw } from 'lucide-svelte';

	let { data } = $props();

	let searchInput = $state('');
	let statusFilter = $state('');

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
			<input
				type="text"
				placeholder="Search hostname, user, name, email..."
				bind:value={searchInput}
				onkeydown={(e) => e.key === 'Enter' && applyFilters()}
				class="w-full rounded-md border border-input bg-background py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
			/>
		</div>
		<select
			bind:value={statusFilter}
			onchange={applyFilters}
			class="rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
		>
			<option value="">All statuses</option>
			<option value="new">New</option>
			<option value="in_review">In Review</option>
			<option value="resolved">Resolved</option>
			<option value="closed">Closed</option>
		</select>
		<button
			onclick={applyFilters}
			class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
		>
			<Filter class="h-4 w-4" />
			Filter
		</button>
		<button
			onclick={refreshReports}
			class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
		>
			<RefreshCcw class="h-4 w-4" />
			Refresh
		</button>
		{#if data.filters.search || data.filters.status}
			<button
				onclick={clearFilters}
				class="rounded-md border border-input px-3 py-2 text-sm text-muted-foreground hover:bg-accent hover:text-foreground"
			>
				Clear
			</button>
		{/if}
	</div>

	<!-- Table -->
	<div class="overflow-hidden rounded-lg border border-border">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-border bg-muted/50">
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">ID</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Hostname</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">User</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Reporter</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Status</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Files</th>
					<th class="px-4 py-3 text-left font-medium text-muted-foreground">Created</th>
				</tr>
			</thead>
			<tbody>
				{#each data.reports as report (report.id)}
					<tr
						class="border-b border-border transition-colors hover:bg-muted/30 cursor-pointer"
						onclick={() => goto(`/reports/${report.id}`)}
					>
						<td class="px-4 py-3 font-mono text-muted-foreground">#{report.id}</td>
						<td class="px-4 py-3">{report.hostname || '—'}</td>
						<td class="px-4 py-3">{report.os_user || '—'}</td>
						<td class="px-4 py-3">
							<div>{report.name}</div>
							<div class="text-xs text-muted-foreground">{report.email}</div>
						</td>
						<td class="px-4 py-3">
							<span
								class="inline-flex rounded-full border px-2 py-0.5 text-xs font-medium {statusColors[report.status]}"
							>
								{statusLabels[report.status]}
							</span>
						</td>
						<td class="px-4 py-3">
							{#if report.file_count > 0}
								<span class="inline-flex items-center gap-1 text-muted-foreground">
									<Paperclip class="h-3.5 w-3.5" />
									{report.file_count}
								</span>
							{:else}
								<span class="text-muted-foreground">—</span>
							{/if}
						</td>
						<td class="px-4 py-3 text-muted-foreground">{formatDate(report.created_at)}</td>
					</tr>
				{:else}
					<tr>
						<td colspan="7" class="px-4 py-12 text-center text-muted-foreground">
							No reports found.
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

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
				<button
					onclick={() => goToPage(data.pagination.page - 1)}
					disabled={data.pagination.page <= 1}
					class="inline-flex items-center rounded-md border border-input p-2 text-sm hover:bg-accent disabled:opacity-50 disabled:pointer-events-none"
				>
					<ChevronLeft class="h-4 w-4" />
				</button>
				{#each Array.from({ length: data.pagination.totalPages }, (_, i) => i + 1) as p}
					{#if p === 1 || p === data.pagination.totalPages || (p >= data.pagination.page - 1 && p <= data.pagination.page + 1)}
						<button
							onclick={() => goToPage(p)}
							class="inline-flex h-9 w-9 items-center justify-center rounded-md text-sm {p ===
							data.pagination.page
								? 'bg-primary text-primary-foreground'
								: 'border border-input hover:bg-accent'}"
						>
							{p}
						</button>
					{:else if p === data.pagination.page - 2 || p === data.pagination.page + 2}
						<span class="px-1 text-muted-foreground">...</span>
					{/if}
				{/each}
				<button
					onclick={() => goToPage(data.pagination.page + 1)}
					disabled={data.pagination.page >= data.pagination.totalPages}
					class="inline-flex items-center rounded-md border border-input p-2 text-sm hover:bg-accent disabled:opacity-50 disabled:pointer-events-none"
				>
					<ChevronRight class="h-4 w-4" />
				</button>
			</div>
		</div>
	{/if}
</div>
