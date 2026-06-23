<script lang="ts">
	import { ModeWatcher } from "mode-watcher";

	import * as Sidebar from "$components/ui/sidebar";
	import AppSidebar from "./app-sidebar/AppSidebar.svelte";
	import PageHeader from "./PageHeader.svelte";

	import { initAppShell } from "$lib/app-shell.svelte";
	import { initUserSessionState } from "$lib/user-session.svelte";
	import { initUserOncallInformationState } from "$lib/userOncall.svelte";
	import { Spinner } from "$src/components/ui/spinner";
	import IncidentCreateDialog, {
		initIncidentCreateDialogController,
	} from "$features/incidents/components/create-incident-dialog";

	const { children } = $props();

	const shell = initAppShell();
	const sess = initUserSessionState();
	initUserOncallInformationState();
	initIncidentCreateDialogController();
</script>

<svelte:head>
	<title>{shell.pageTitle}</title>
</svelte:head>

<ModeWatcher />

{#if sess.ready}
	<Sidebar.Provider>
		{#if sess.isSetup}
			<AppSidebar variant="sidebar" />
		{/if}
		<main class="antialiased flex flex-col flex-1 min-w-0 min-h-0 h-dvh overflow-hidden">
			{#if sess.isSetup}
				<div class="flex w-full justify-between items-center h-14 bg-surface-200 border-b px-2">
					<PageHeader />
				</div>
			{/if}

			<div id="scroll-body" class="flex-1 flex min-h-0 overflow-y-auto p-2">
				{@render children()}
			</div>
		</main>
		<IncidentCreateDialog />
	</Sidebar.Provider>
{:else}
	<div class="w-full h-dvh grid place-items-center">
		<Spinner />
	</div>
{/if}
