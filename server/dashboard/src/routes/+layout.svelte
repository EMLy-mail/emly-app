<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { Bug, LayoutDashboard } from 'lucide-svelte';

	let { children, data } = $props();
</script>

<div class="flex h-screen overflow-hidden">
	<!-- Sidebar -->
	<aside class="flex w-56 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
		<div class="flex items-center gap-2 border-b border-sidebar-border px-4 py-4">
			<Bug class="h-6 w-6 text-primary" />
			<span class="text-lg font-semibold">EMLy Dashboard</span>
		</div>
		<nav class="flex-1 px-2 py-3">
			<a
				href="/"
				class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors {$page.url
					.pathname === '/'
					? 'bg-accent text-accent-foreground'
					: 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'}"
			>
				<LayoutDashboard class="h-4 w-4" />
				Reports
			</a>
		</nav>
		<div class="border-t border-sidebar-border px-4 py-3 text-xs text-muted-foreground">
			EMLy Bug Reports
		</div>
	</aside>

	<!-- Main content -->
	<div class="flex flex-1 flex-col overflow-hidden">
		<!-- Top bar -->
		<header
			class="flex h-14 items-center justify-between border-b border-border bg-card px-6"
		>
			<h1 class="text-lg font-semibold">
				{#if $page.url.pathname === '/'}
					Bug Reports
				{:else if $page.url.pathname.startsWith('/reports/')}
					Report Detail
				{:else}
					Dashboard
				{/if}
			</h1>
			{#if data.newCount > 0}
				<div class="flex items-center gap-2 rounded-md bg-blue-500/10 px-3 py-1.5 text-sm text-blue-400">
					<span class="inline-block h-2 w-2 rounded-full bg-blue-400"></span>
					{data.newCount} new {data.newCount === 1 ? 'report' : 'reports'}
				</div>
			{/if}
		</header>

		<!-- Page content -->
		<main class="flex-1 overflow-auto p-6">
			{@render children()}
		</main>
	</div>
</div>
