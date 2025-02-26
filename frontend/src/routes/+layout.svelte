<script lang="ts">
	import "$src/app.postcss";
	import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
	import AppShell from "$features/app/views/app-shell/AppShell.svelte";
	import { browser, dev } from "$app/environment";

	const { children } = $props();

	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				enabled: browser,
				retry: false,
				refetchOnWindowFocus: dev,
				staleTime: 5000,
			},
		},
	});
</script>

<svelte:head>
	<title>Rezible</title>
</svelte:head>

<QueryClientProvider client={queryClient}>
	<AppShell>{@render children()}</AppShell>
</QueryClientProvider>
