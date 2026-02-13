<script lang="ts">
  	import { ModeWatcher } from "mode-watcher";
	import { AuthSessionState, setAuthSessionState } from "$lib/auth.svelte";
	import { settings } from "$lib/settings.svelte";
	import { setUserOncallInformationState } from "$lib/userOncall.svelte";
	import { setToastState } from "$lib/toasts.svelte";
	import { appShell } from "$features/app";
	import Sidebar from "./sidebar/Sidebar.svelte";
	import Toaster from "../toaster/Toaster.svelte";
	import PageContainer from "./PageContainer.svelte";
	import SessionProtector from "./SessionProtector.svelte";

	const { children } = $props();
	
	const session = new AuthSessionState();
	setAuthSessionState(session);

	appShell.setup();
	settings.setup();

	setToastState();
	setUserOncallInformationState();
</script>

<svelte:head>
	<title>Rezible</title>
</svelte:head>

<ModeWatcher />

<Toaster />

<div class="antialiased p-2 flex gap-2 w-dvw h-dvh min-h-dvh bg-surface-300 text-surface-content">
	<SessionProtector>
		{#if session.isAuthenticated && session.isSetup}
			<Sidebar />
		{/if}

		<main class="w-full max-w-full h-full max-h-full min-h-0 flex flex-col">
			<PageContainer>
				{@render children()}
			</PageContainer>
		</main>
	</SessionProtector>
</div>
