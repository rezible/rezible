<script lang="ts">
	import "$src/app.postcss";
	import { QueryClientProvider } from "@tanstack/svelte-query";
	import { appState } from "$lib/appState.svelte";
	import { session } from "$lib/auth.svelte";
	import AppShell from "$features/app/views/app-shell/AppShell.svelte";
	import AuthSessionErrorView from "$features/app/views/auth-session-error/AuthSessionErrorView.svelte";

	const { data, children } = $props();

	appState.setup();
</script>

<svelte:head>
	<title>Rezible</title>
</svelte:head>

<QueryClientProvider client={data.queryClient}>
	{#if session.error}
		<AuthSessionErrorView />
	{:else}
		<AppShell>{@render children()}</AppShell>
	{/if}
</QueryClientProvider>
