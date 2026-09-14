<script lang="ts" module>
	import { resolve } from "$app/paths";
	import type { RouteId } from "$app/types";
	import type { Component } from "svelte";

	export type FeatureNavRailEntry<Route extends RouteId> = {
		label: string;
		params: Parameters<typeof resolve<Route>>[1];
		component: Component;
		icon?: Component;
	};
</script>

<script lang="ts" generics="Route extends RouteId">
	import { page } from "$app/state";
	import { useAppShell } from "$lib/app-shell.svelte";
	import type { ResolvedPathname } from "$app/types";
	import type { Snippet } from "svelte";

	type Props = {
		route: Route;
		entries: FeatureNavRailEntry<Route>[];
		label: string;
		featureLinks?: Snippet;
	};
	const { route, entries, label, featureLinks }: Props = $props();

	useAppShell().registerFeatureRail();

	const resolveEntryRoute = resolve as unknown as (
		route: Route,
		params: FeatureNavRailEntry<Route>["params"]
	) => ResolvedPathname;

	const paths = $derived(entries.map((entry) => resolveEntryRoute(route, entry.params)));
	const activeIndex = $derived.by(() => {
		if (page.route.id !== route) return undefined;
		return paths.findIndex((path) => page.url.pathname === path);
	});
	const activePath = $derived(activeIndex !== undefined && activeIndex >= 0 ? paths[activeIndex] : undefined);
	const ActiveComponent = $derived(
		activeIndex !== undefined && activeIndex >= 0 ? entries[activeIndex].component : undefined
	);
</script>

<div class="flex h-full min-h-0 min-w-0 flex-1 flex-col overflow-hidden bg-background md:flex-row">
	<aside
		class="bg-feature text-feature-foreground flex min-h-0 w-full min-w-0 shrink-0 flex-col overflow-hidden border-b md:w-[180px] md:border-e md:border-b-0"
	>
		<nav aria-label={label} class="flex shrink-0 flex-row gap-1 overflow-x-auto md:flex-col md:overflow-visible">
			{#each entries as entry, index (index)}
				{@const href = paths[index]}
				{@const isActive = activePath === href}
				<a 
					data-slot="feature-navigation-link"
					{href}
					aria-current={isActive ? "page" : undefined}
					data-active={isActive ? "true" : undefined}
					class="hover:bg-accent focus-visible:outline-ring data-[active=true]:bg-selection data-[active=true]:text-selection-foreground relative flex h-9 min-w-0 shrink-0 items-center gap-2 px-3 text-sm font-medium text-muted-foreground outline-none hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-2 after:bg-brand after:absolute after:inset-x-2 after:bottom-0 after:h-0.5 after:opacity-0 data-[active=true]:after:opacity-100 md:after:inset-x-auto md:after:inset-y-0 md:after:right-0 md:after:h-auto md:after:w-0.5"
				>
					{#if entry.icon}
						<entry.icon class="size-4 shrink-0" aria-hidden="true" />
					{/if}
					<span class="truncate">{entry.label}</span>
				</a>
			{/each}
		</nav>

		{#if featureLinks}
			<div class="min-h-0 flex-1 overflow-y-auto border-t border-border p-3">
				{@render featureLinks()}
			</div>
		{/if}
	</aside>
	
	<section class="flex min-h-0 min-w-0 flex-1 flex-col overflow-y-auto p-4">
		{#if ActiveComponent}
			<ActiveComponent />
		{/if}
	</section>
</div>
