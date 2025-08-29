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

	const routeId = $derived(page.route.id);

	const AuthRouteId = "/auth";
	const isAuthRoute = $derived(!!routeId?.startsWith(AuthRouteId));

	const orgNeedsSetup = $derived(session.isAuthenticated && !!session.org?.requiresInitialSetup);
	const SetupRouteId = "/setup";
	const isSetupRoute = $derived(!!routeId?.startsWith(SetupRouteId));

	const redirectTo = $derived.by(() => {
		if (!session.loaded) return;
		if (!session.isAuthenticated && !isAuthRoute) return AuthRouteId;
		if (orgNeedsSetup && !isSetupRoute) return SetupRouteId;
		if (session.isAuthenticated && isAuthRoute) return "/";
	});
	const navigatingTo = $derived(navigating.to?.route.id);

	watch(() => redirectTo, route => {
		if (!!route && route !== navigatingTo) goto(route);
	});

	const showPage = $derived(session.loaded && !redirectTo);
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