<script lang="ts">
  	import { ModeWatcher } from "mode-watcher";
	import { initAuthSessionState } from "$lib/auth.svelte";
	import { initUserOncallInformationState } from "$lib/userOncall.svelte";
	import { initToastState } from "$lib/toasts.svelte";
	import { appShell } from "$features/app";
	import Sidebar from "./sidebar/Sidebar.svelte";
	import Toaster from "../toaster/Toaster.svelte";
	import PageContainer from "./PageContainer.svelte";
	import Header from "./header/Header.svelte";
	import LoadingIndicator from "$src/components/loading-indicator/LoadingIndicator.svelte";

	const { children } = $props();
	
	const session = initAuthSessionState();

	appShell.setup();

	initToastState();
	initUserOncallInformationState();
</script>

<svelte:head>
	<title>{appShell.pageTitle}</title>
</svelte:head>

<ModeWatcher />

<Toaster />

<div class="antialiased flex flex-col overflow-hidden w-dvw h-dvh min-h-dvh bg-surface-300 text-surface-content">
	{#if !session.ready}
		<div class="w-full h-full grid place-items-center">
			<LoadingIndicator />
		</div>
	{:else}
		<Header />

		<div class="flex flex-1 min-h-0 overflow-hidden">
			{#if session.isSetup}
				<Sidebar />
			{/if}

			<PageContainer>
				{@render children()}
			</PageContainer>
		</div>
	{/if}
</div>