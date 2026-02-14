<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="flex min-h-[60vh] flex-col items-center justify-center text-center">
	<h1 class="text-6xl font-bold text-muted-foreground">{$page.status}</h1>
	<p class="mt-4 text-lg text-muted-foreground">{$page.error?.message || 'Something went wrong'}</p>
	<a
		href="/"
		class="mt-6 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
	>
		Back to Reports
	</a>
</div>
