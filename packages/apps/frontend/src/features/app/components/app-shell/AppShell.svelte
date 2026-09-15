<script lang="ts">
	import { ModeWatcher } from "mode-watcher";

	import { initAppShell } from "$lib/app-shell.svelte";
	import { initUserSessionState } from "$lib/user-session.svelte";
	import { initUserOncallInformationState } from "$lib/userOncall.svelte";
	import { cn } from "$lib/utils";

	import * as Sidebar from "$components/ui/sidebar";
	import { Spinner } from "$components/ui/spinner";
	import AppSidebar from "./app-sidebar/AppSidebar.svelte";
	import PageHeader from "./PageHeader.svelte";

	const { children } = $props();

	const shell = initAppShell();
	const sess = initUserSessionState();

	// TODO: remove this
	initUserOncallInformationState();

	const pageDescriptor = $derived(shell.pageDescriptor);
	const featureRailActive = $derived(shell.featureRail);
</script>

<svelte:head>
	<title>{pageDescriptor ? `Rezible - ${pageDescriptor.title}` : "Rezible"}</title>
</svelte:head>

<ModeWatcher />

{#if sess.ready}
	<Sidebar.Provider {featureRailActive}>
		{#if sess.isSetup}
			<AppSidebar variant="sidebar" />
		{/if}
		<main class="flex h-dvh min-h-0 min-w-0 flex-1 flex-col overflow-hidden antialiased">
			{#if sess.isSetup}
				<div class="bg-card flex h-14 w-full items-center justify-between border-b px-4">
					{#if pageDescriptor}
						<PageHeader {pageDescriptor} {featureRailActive} />
					{/if}
				</div>
			{/if}

			<div id="scroll-body" class={cn("flex min-h-0 flex-1", featureRailActive ? "overflow-hidden" : "overflow-y-auto")}>
				{@render children()}
			</div>
		</main>
	</Sidebar.Provider>
{:else}
	<div class="w-full h-dvh grid place-items-center">
		<Spinner />
	</div>
{/if}
