<script lang="ts">
	import { ModeWatcher } from "mode-watcher";

	import * as Sidebar from "$components/ui/sidebar";
	import AppSidebar from "./app-sidebar/AppSidebar.svelte";
	import PageHeader from "./PageHeader.svelte";

	import { initAppShell } from "$lib/app-shell.svelte";
	import { initAuthSessionState } from "$src/lib/auth-session.svelte";
	import { initUserOncallInformationState } from "$lib/userOncall.svelte";
	import { Spinner } from "$src/components/ui/spinner";

	const { children } = $props();

	const shell = initAppShell();
	const auth = initAuthSessionState();
	initUserOncallInformationState();
</script>

<svelte:head>
	<title>{shell.pageTitle}</title>
</svelte:head>

<ModeWatcher />

{#if auth.ready}
	<Sidebar.Provider>
		{#if auth.isSetup}
			<AppSidebar variant="sidebar" />
		{/if}
		<main class="antialiased flex flex-col flex-1 min-w-0 min-h-0 h-dvh overflow-hidden">
			{#if auth.isSetup}
				<div class="flex w-full justify-between items-center h-11 bg-surface-200 border-b px-1">
					<PageHeader />
				</div>
			{/if}

			<div id="scroll-body" class="flex-1 min-h-0 overflow-y-auto p-2">
				{@render children()}
			</div>
		</main>
	</Sidebar.Provider>
{:else}
	<div class="w-full h-dvh grid place-items-center">
		<Spinner />
	</div>
{/if}
