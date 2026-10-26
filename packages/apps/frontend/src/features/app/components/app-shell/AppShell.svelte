<script lang="ts">
	import { ModeWatcher } from "mode-watcher";

	import { initAppShell } from "$lib/app-shell.svelte";
	import { initUserSessionState } from "$lib/user-session.svelte";
	import { cn } from "$lib/utils";

	import * as Sidebar from "$components/ui/sidebar";
	import CreateIncidentDialog, {
		initIncidentCreateDialogController,
	} from "$features/incidents/components/create-incident-dialog";
	import { Spinner } from "$components/ui/spinner";
	import AppSidebar from "./app-sidebar/AppSidebar.svelte";
	import PageHeader from "./PageHeader.svelte";

	const { children } = $props();

	const shell = initAppShell();
	const sess = initUserSessionState();
	// Declaring an incident is offered from several pages, so its dialog is shared.
	initIncidentCreateDialogController();

	const pageDescriptor = $derived(shell.pageDescriptor);
	const documentTitle = $derived.by(() => {
		if (!pageDescriptor) {
			return "Rezible";
		}
		if (pageDescriptor.reference) {
			return `Rezible - ${pageDescriptor.reference} · ${pageDescriptor.title}`;
		}
		return `Rezible - ${pageDescriptor.title}`;
	});
	const featureRailActive = $derived(shell.featureRail);
</script>

<svelte:head>
	<title>{documentTitle}</title>
</svelte:head>

<ModeWatcher />

{#if sess.ready}
	<Sidebar.Provider collapsed={featureRailActive}>
		{#if sess.isSetup}
			<AppSidebar variant="sidebar" />
		{/if}
		<main class="flex h-dvh min-h-0 min-w-0 flex-1 flex-col overflow-hidden antialiased">
			{#if sess.isSetup && !!pageDescriptor}
				<PageHeader descriptor={pageDescriptor} />
			{/if}

			<div
				id="scroll-body"
				class={cn("flex min-h-0 flex-1", featureRailActive ? "overflow-hidden" : "overflow-y-auto")}
			>
				{@render children()}
			</div>
		</main>
	</Sidebar.Provider>
	{#if sess.isSetup}
		<CreateIncidentDialog />
	{/if}
{:else}
	<div class="w-full h-dvh grid place-items-center">
		<Spinner />
	</div>
{/if}
