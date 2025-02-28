<script lang="ts">
	import "$src/app.postcss";
	import { QueryClientProvider } from "@tanstack/svelte-query";
	import AppShell from "$features/app/views/app-shell/AppShell.svelte";

	const { data, children } = $props();
</script>

<svelte:head>
	<title>Rezible</title>
</svelte:head>

<QueryClientProvider client={data.queryClient}>
	<AppShell>{@render children()}</AppShell>
</QueryClientProvider>
