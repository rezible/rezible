<script lang="ts">
  	import { ModeWatcher } from "mode-watcher";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";

	import * as Sidebar from "$components/ui/sidebar";
	import AppSidebar from "./app-sidebar/AppSidebar.svelte";
	import PageHeader from "./PageHeader.svelte";

	import { initAppShell } from "$lib/appShell.svelte";
	import { initAuthSessionState } from "$lib/auth.svelte";
	import { initUserOncallInformationState } from "$lib/userOncall.svelte";

	const { children } = $props();
	
	const shell = initAppShell();
    const auth = initAuthSessionState();
    initUserOncallInformationState();
</script>

<svelte:head>
	<title>{shell.pageTitle}</title>
</svelte:head>

<ModeWatcher />

<Sidebar.Provider>
	{#if auth.isSetup}
		<AppSidebar variant="sidebar" />
	{/if}
	<main class="antialiased flex flex-col flex-1 min-w-0 min-h-0 h-dvh overflow-hidden">
		{#if !auth.ready}
			<div class="flex-1 grid place-items-center">
				<LoadingIndicator />
			</div>
		{:else}
			{#if auth.isSetup}
				<div class="flex w-full justify-between items-center h-11 rounded-md bg-surface-200 border-b">
					<PageHeader />
				</div>
			{/if}

			<div id="scroll-body" class="flex-1 min-h-0 overflow-y-auto p-2">
				{@render children()}
			</div>
		{/if}
	</main>
</Sidebar.Provider>