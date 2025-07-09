<script lang="ts">
	import "$src/app.postcss";
	import { QueryClientProvider } from "@tanstack/svelte-query";
	import { ThemeInit } from "svelte-ux";
	import AppShell from "$features/app/views/app-shell/AppShell.svelte";

	const { data, children } = $props();
</script>

<svelte:head>
	<title>Rezible</title>
</svelte:head>

<ThemeInit />

<QueryClientProvider client={data.queryClient}>
	<AppShell>{@render children()}</AppShell>
</QueryClientProvider>
