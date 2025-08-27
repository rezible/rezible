<script lang="ts">
	import { AuthSessionState, setAuthSessionState } from "$lib/auth.svelte";
	import { settings } from "$lib/settings.svelte";
	import { setUserOncallInformationState } from "$lib/userOncall.svelte";
	import { setToastState } from "$lib/toasts.svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import Sidebar from "./sidebar/Sidebar.svelte";
	import Toaster from "./toaster/Toaster.svelte";
	import PageContainer from "./PageContainer.svelte";
	import { navigating, page } from "$app/state";
	import { goto } from "$app/navigation";
	import { watch } from "runed";

	const { children } = $props();
	
	const session = new AuthSessionState();
	setAuthSessionState(session);

	appShell.setup();
	settings.setup();

	setToastState();
	setUserOncallInformationState();

	const isAuthError = $derived(!!session.error);
	const isAuthRoute = $derived(!!page.route.id?.startsWith("/auth"));

	const authRedirectPath = $derived.by(() => {
		if (!session.loaded) return;
		if (isAuthError && !isAuthRoute) return "/auth";
		if (!isAuthError && isAuthRoute) return "/";
	});
	const navigatingTo = $derived(navigating.to?.route.id);

	watch(() => authRedirectPath, path => {
		if (!path) return;
		if (path === navigatingTo) return;
		goto(path);
	});

	const showPage = $derived(session.loaded && !authRedirectPath);
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-300 text-surface-content">
	{#if session.isAuthenticated}
		<Sidebar />
	{/if}

	<main class="w-full h-full p-2">
		<PageContainer hideNavBar={!session.isAuthenticated}>
			{#if showPage}
				{@render children()}
			{/if}
		</PageContainer>
	</main>
</div>

<Toaster />