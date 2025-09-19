<script lang="ts">
	import { AuthSessionState, setAuthSessionState } from "$lib/auth.svelte";
	import { settings } from "$lib/settings.svelte";
	import { setUserOncallInformationState } from "$lib/userOncall.svelte";
	import { setToastState } from "$lib/toasts.svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import Sidebar from "./sidebar/Sidebar.svelte";
	import Toaster from "./toaster/Toaster.svelte";
	import PageContainer from "./PageContainer.svelte";
	import SessionProtector from "./SessionProtector.svelte";
	import LogoHeader from "./sidebar/LogoHeader.svelte";

	const { children } = $props();
	
	const session = new AuthSessionState();
	setAuthSessionState(session);

	appShell.setup();
	settings.setup();

	setToastState();
	setUserOncallInformationState();
</script>

<div class="antialiased w-dvw h-dvh min-h-dvh bg-surface-300 text-surface-content">
	<SessionProtector>
		{#if session.isAuthenticated}
			{#if session.isSetup}
				<Sidebar />
			{:else}
				<div class="w-full">
					<LogoHeader />
				</div>
			{/if}
		{/if}

		<main class="w-full max-w-full h-full max-h-full min-h-0 p-2 flex flex-col">
			<PageContainer>
				{@render children()}
			</PageContainer>
		</main>
	</SessionProtector>
</div>

<Toaster />