<script lang="ts">
	import "$src/app.postcss";
	import { browser, dev } from "$app/environment";
	import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
	import { ThemeInit } from "svelte-ux";
	import { AppShellView } from "$features/app-shell";

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

<ThemeInit />

<QueryClientProvider client={queryClient}>
	<AppShellView>{@render children()}</AppShellView>
</QueryClientProvider>
