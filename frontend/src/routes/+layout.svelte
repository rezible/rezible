<script lang="ts">
	import "$src/app.postcss";
	import { settings } from "svelte-ux";
	import { QueryClientProvider } from "@tanstack/svelte-query";
	import { session } from "$lib/auth.svelte";
	import AppShell from "$features/app/views/app-shell/AppShell.svelte";
	import AuthSessionErrorView from "$features/app/views/auth-session-error/AuthSessionErrorView.svelte";

	const { data, children } = $props();

	// TODO: dictionary
	settings({
		themes: { light: ["light-old"], dark: ["dark", "bleh"] },
	});
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
