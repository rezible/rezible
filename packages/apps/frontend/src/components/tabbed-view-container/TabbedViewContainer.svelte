<script lang="ts" module>
	import { resolve } from "$app/paths";

	export type Tab<Route extends RouteId> = {
		label: string;
		params: (Parameters<typeof resolve<Route>>)[1];
		component: Component;
	};
</script>

<script lang="ts" generics="Route extends RouteId">
	import { page } from "$app/state";
	import type { Pathname } from "$app/types";
	import type { ResolvedPathname } from "$app/types";
	import type { RouteId } from "$app/types";
	import type { Component, Snippet } from "svelte";
	import { SvelteMap } from "svelte/reactivity";

	type Props = { 
		route: Route;
		tabs: Tab<Route>[];
		infoBar?: Snippet;
		tabSidebar?: Snippet;
	};
	const { route, tabs, infoBar, tabSidebar }: Props = $props();

	type ResolveParams = Parameters<typeof resolve<Route>>;
	const tabPaths = $derived<ResolvedPathname[]>(tabs.map(t => resolve(...([route, t.params] as ResolveParams))));
	
	const activeTab = $derived.by(() => {
		const currRoute = page.route.id;
		const currPath = page.url.pathname;
		if (!currRoute || currRoute !== route) return;
		return tabs.find((t, i) => (currPath === tabPaths.at(i)));
	});
	
	const ActiveComponent = $derived(activeTab?.component);
</script>

<div class="flex-1 flex flex-col h-full max-h-full min-h-0 overflow-auto">
	<div class="w-full flex h-12 z-[1] justify-between">
		<div class="flex gap-1 self-end">
			{#each tabs as tab, i}
				<a href="{tabPaths[i]}" data-active={(tab.label === activeTab?.label) ? true : undefined}
					class="group inline-flex self-end h-12 p-4 py-3 text-lg border border-surface-100 border-b-0 rounded-t-lg relative text-muted-foreground data-active:bg-surface-200 data-active:text-foreground">
					<span class="leading-none self-center">
						{tab.label}
					</span>
					<div class="bottom-0 left-0 -mb-px w-full border-b border-surface-200 absolute hidden group-data-[active]:block"></div>
				</a>
			{/each}
		</div>

		{#if !!infoBar}
			<div class="flex gap-4 h-12 max-h-14 overflow-y-hidden justify-between pb-1">
				{@render infoBar()}
			</div>
		{/if}
	</div>

	<div class="flex-1 flex border border-surface-100 rounded-b-lg rounded-tr-lg bg-surface-200 overflow-y-auto">
		<div class="flex-1 min-h-0 max-h-full overflow-y-auto p-2">
			<ActiveComponent />
		</div>
		{@render tabSidebar?.()}
	</div>
</div>
