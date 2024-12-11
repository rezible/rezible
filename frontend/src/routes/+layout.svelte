<script lang="ts">
	import '$src/app.postcss';
	import { settings } from 'svelte-ux';
	import { QueryClientProvider, createQuery } from '@tanstack/svelte-query';
	import AppShell from '$features/app/views/app-shell/AppShell.svelte';
	import themes from "$src/themes.json";
    import { listUserNotificationsOptions } from '$lib/api/index.js';

	const { data, children } = $props();

	// TODO: dictionary
	settings({
		themes: {light: ["light-old"], dark: ["dark", "bleh"]}
	});
</script>

<svelte:head>
	<title>Rezible</title>
</svelte:head>

<QueryClientProvider client={data.queryClient}>
	<AppShell>
		{@render children()}
	</AppShell>
</QueryClientProvider>
