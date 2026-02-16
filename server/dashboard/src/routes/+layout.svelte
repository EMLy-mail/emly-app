<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { enhance } from '$app/forms';
	import { onMount, onDestroy } from 'svelte';
	import { Bug, LayoutDashboard, Users, LogOut } from 'lucide-svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Separator } from '$lib/components/ui/separator';
	import { presence } from '$lib/stores/presence.svelte';

	let { children, data } = $props();

	const otherActiveUsers = $derived(
		presence.activeUsers.filter((u) => u.userId !== data.user?.id)
	);

	onMount(() => {
		if (data.user) {
			presence.connect();
		}
	});

	onDestroy(() => {
		presence.disconnect();
	});
</script>

{#if !data.user}
	{@render children()}
{:else}
	<Sidebar.Provider>
		<Sidebar.Root>
			<Sidebar.Header>
				<div class="flex items-center justify-center border-b border-border p-3" style="padding: 12px; display: flex; justify-content: center;">
					<Bug class="h-6 w-6 text-primary" />
					<span class="mt-2 pl-3 text-lg font-bold">EMLy Dashboard</span>
				</div>
			</Sidebar.Header>
			<Sidebar.Content>
				<Sidebar.Group>
					<Sidebar.GroupLabel>Menu</Sidebar.GroupLabel>
					<Sidebar.GroupContent>
						<Sidebar.Menu>
							<Sidebar.MenuItem>
								<Sidebar.MenuButton isActive={$page.url.pathname === '/'}>
									{#snippet child({ props })}
										<a href="/" {...props}>
											<LayoutDashboard />
											<span>Reports</span>
										</a>
									{/snippet}
								</Sidebar.MenuButton>
							</Sidebar.MenuItem>
							{#if data.user.role === 'admin'}
								<Sidebar.MenuItem>
									<Sidebar.MenuButton isActive={$page.url.pathname === '/users'}>
										{#snippet child({ props })}
											<a href="/users" {...props}>
												<Users />
												<span>Users</span>
											</a>
										{/snippet}
									</Sidebar.MenuButton>
								</Sidebar.MenuItem>
							{/if}
						</Sidebar.Menu>
					</Sidebar.GroupContent>
				</Sidebar.Group>
			</Sidebar.Content>
			<Sidebar.Footer>
				<Sidebar.Menu>
					<Sidebar.MenuItem>
						<Sidebar.MenuButton
							size="lg"
							class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
						>
							{#snippet child({ props })}
								<div {...props}>
									<div class="grid flex-1 text-left text-sm leading-tight">
										<span class="pb-1 truncate font-semibold">{data.user.displayname || data.user.username}</span>
										<span
											class="w-fit inline-flex rounded-full border px-2 py-0.5 text-xs font-medium truncate {data.user.role ===
											'admin'
												? 'bg-purple-500/20 text-purple-400 border-purple-500/30'
												: 'bg-zinc-500/20 text-zinc-400 border-zinc-500/30'}"
										>
											{data.user.role}
										</span>
									</div>
									<form method="POST" action="/logout" use:enhance>
										<button type="submit" title="Sign out">
											<LogOut class="ml-auto size-4" />
										</button>
									</form>
								</div>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				</Sidebar.Menu>
			</Sidebar.Footer>
		</Sidebar.Root>
		<Sidebar.Inset>
			<header
				class="flex h-16 shrink-0 items-center justify-between border-b px-4 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12"
			>
				<div class="flex flex-1 items-center gap-2 px-4">
					<Sidebar.Trigger class="-ml-1" />
					<Separator orientation="vertical" class="mr-2 h-4" />
					<h1 class="text-lg font-semibold">
						{#if $page.url.pathname === '/'}
							Bug Reports
						{:else if $page.url.pathname.startsWith('/reports/')}
							Report Detail
						{:else if $page.url.pathname === '/users'}
							User Management
						{:else}
							Dashboard
						{/if}
					</h1>
				</div>
				<div class="flex items-center gap-2 px-4">
					{#if otherActiveUsers.length > 0}
						<div class="flex items-center gap-1.5">
							{#each otherActiveUsers.slice(0, 5) as activeUser}
								<Tooltip.Root>
									<Tooltip.Trigger>
										<div class="relative flex h-7 w-7 items-center justify-center rounded-full bg-green-500/20 text-xs font-medium text-green-400 border border-green-500/30">
											{(activeUser.displayname || activeUser.username).charAt(0).toUpperCase()}
											<span class="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full bg-green-500 border-2 border-background"></span>
										</div>
									</Tooltip.Trigger>
									<Tooltip.Content>
										<p class="font-medium">{activeUser.displayname || activeUser.username}</p>
										<p class="text-xs text-muted-foreground">
											{#if activeUser.reportId}
												Viewing Report #{activeUser.reportId}
											{:else if activeUser.currentPath === '/users'}
												User Management
											{:else if activeUser.currentPath === '/'}
												Reports List
											{:else}
												{activeUser.currentPath}
											{/if}
										</p>
									</Tooltip.Content>
								</Tooltip.Root>
							{/each}
							{#if otherActiveUsers.length > 5}
								<span class="text-xs text-muted-foreground">+{otherActiveUsers.length - 5}</span>
							{/if}
						</div>
					{/if}
					{#if data.newCount > 0}
						<div
							class="ml-4 flex items-center gap-2 rounded-md bg-blue-500/10 px-3 py-1.5 text-sm text-blue-400"
						>
							<span class="inline-block h-2 w-2 rounded-full bg-blue-400"></span>
							{data.newCount} new {data.newCount === 1 ? 'report' : 'reports'}
						</div>
					{/if}
				</div>
			</header>
			<div class="flex flex-1 flex-col gap-4 p-4 pt-0 mt-2">
				{@render children()}
			</div>
		</Sidebar.Inset>
	</Sidebar.Provider>
{/if}
