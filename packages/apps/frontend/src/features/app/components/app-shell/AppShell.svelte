<script lang="ts">
  	import { ModeWatcher } from "mode-watcher";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import Header from "./header/Header.svelte";
	import Sidebar from "./sidebar/Sidebar.svelte";
	import PageContainer from "./PageContainer.svelte";

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

<div class="antialiased flex flex-col overflow-hidden w-dvw h-dvh min-h-dvh bg-surface-300 text-surface-content">
	{#if !auth.ready}
		<div class="w-full h-full grid place-items-center">
			<LoadingIndicator />
		</div>
	{:else}
		<Header />

		<div class="flex flex-1 min-h-0 overflow-hidden">
			{#if auth.isSetup}
				<Sidebar />
			{/if}

			<PageContainer>
				{@render children()}
			</PageContainer>
		</div>
	{/if}
</div>