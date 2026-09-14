<script lang="ts">
	import { ModeWatcher } from "mode-watcher";

	import * as Sidebar from "$components/ui/sidebar";
	import AppSidebar from "./app-sidebar/AppSidebar.svelte";
	import PageHeader from "./PageHeader.svelte";

	import { initAppShell } from "$lib/app-shell.svelte";
	import { initUserSessionState } from "$lib/user-session.svelte";
	import { initUserOncallInformationState } from "$lib/userOncall.svelte";
	import { Spinner } from "$src/components/ui/spinner";

	const { children } = $props();

	const shell = initAppShell();
	const sess = initUserSessionState();
	initUserOncallInformationState();
</script>

<svelte:head>
	<title>{shell.pageDescriptor ? `${shell.pageDescriptor.title} · Rezible` : "Rezible"}</title>
</svelte:head>

<ModeWatcher />

{#if sess.ready}
	<Sidebar.Provider viewRail={shell.viewRail}>
		{#if sess.isSetup}
			<AppSidebar variant="sidebar" />
		{/if}
		<main class="flex h-dvh min-h-0 min-w-0 flex-1 flex-col overflow-hidden antialiased">
			{#if sess.isSetup}
				<div class="bg-card flex h-14 w-full items-center justify-between border-b px-6">
					<PageHeader />
				</div>
			{/if}

			<div
				id="scroll-body"
				class:overflow-hidden={shell.viewRail}
				class:overflow-y-auto={!shell.viewRail}
				class:p-6={!shell.viewRail}
				class="flex min-h-0 flex-1"
			>
				{@render children()}
			</div>
		</main>
	</Sidebar.Provider>
{:else}
	<div class="w-full h-dvh grid place-items-center">
		<Spinner />
	</div>
{/if}
