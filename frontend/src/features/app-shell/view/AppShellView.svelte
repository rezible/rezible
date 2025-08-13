<script lang="ts">
	import { session } from "$lib/auth.svelte";
	import { settings } from "$lib/settings.svelte";
	import { setUserOncallInformationState } from "$lib/userOncall.svelte";
	import { setToastState } from "$lib/toasts.svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import Sidebar from "./sidebar/Sidebar.svelte";
	import Toaster from "./toaster/Toaster.svelte";
	import PageContainer from "./PageContainer.svelte";

	const { children } = $props();

	appShell.setup();
	settings.setup();

	setToastState();
	setUserOncallInformationState();

	const isAuthenticated = $derived(!!session.user);
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-300 text-surface-content">
	{#if isAuthenticated}
		<Sidebar />
	{/if}

	<main class="w-full h-full p-2">
		<PageContainer hideNavBar={!isAuthenticated}>
			{@render children()}
		</PageContainer>
	</main>
</div>

<Toaster />