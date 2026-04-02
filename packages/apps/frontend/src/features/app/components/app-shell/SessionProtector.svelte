<script lang="ts">
    import type { Snippet } from "svelte";
	import { watch } from "runed";
	import { goto } from "$app/navigation";
	import { navigating, page } from "$app/state";
	import { useAuthSessionState } from "$lib/auth.svelte";
	import { APP_AUTH_ROUTE_BASE } from "$lib/config";

    const { children }: { children: Snippet } = $props();

    const session = useAuthSessionState();

	const routeId = $derived(page.route.id);

	const isAuthRoute = $derived(!!routeId?.startsWith(APP_AUTH_ROUTE_BASE));

	const SetupRouteId = "/setup";
	const isSetupRoute = $derived(!!routeId?.startsWith(SetupRouteId));

	const redirectTo = $derived.by(() => {
		if (!session.loaded) return;
        if (session.isAuthenticated && isAuthRoute) return "/";
        if (session.isSetup && isSetupRoute) return "/";
		if (!session.isAuthenticated && !isAuthRoute) return APP_AUTH_ROUTE_BASE;
        if (session.isAuthenticated && !session.isSetup && !isSetupRoute) return SetupRouteId;
	});
	const navigatingTo = $derived(navigating.to?.route.id);

	watch(() => redirectTo, route => {
		if (!!route && route !== navigatingTo) goto(route);
	});
</script>

{#if session.loaded && !redirectTo}
    {@render children()}
{/if}