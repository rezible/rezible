<script lang="ts">
	import "$src/app.css";
	import { QueryClientProvider } from "@tanstack/svelte-query";
	import { AppShellView } from "$features/app-shell";
	import type { LayoutProps } from "./$types";

	const { data, children }: LayoutProps = $props();
</script>

<svelte:head>
	<title>Rezible</title>
</svelte:head>

<QueryClientProvider client={data.queryClient}>
	<AppShellView>{@render children()}</AppShellView>
</QueryClientProvider>
