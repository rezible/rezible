<script lang="ts">
	import "$src/app.css";
	import type { LayoutProps } from "./$types";
	import { QueryClientProvider } from "@tanstack/svelte-query";

	import { AppShell } from "$features/app";

	const { data, children }: LayoutProps = $props();
</script>

<QueryClientProvider client={data.queryClient}>
	<AppShell>{@render children()}</AppShell>
</QueryClientProvider>
