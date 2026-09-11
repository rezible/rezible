<script lang="ts" module>
	import { resolve } from "$app/paths";
	import type { RouteId } from "$app/types";
	import type { Component } from "svelte";

	export type ViewRailEntry<Route extends RouteId> = {
		label: string;
		params: Parameters<typeof resolve<Route>>[1];
		component: Component;
		icon?: Component;
	};
</script>

<script lang="ts" generics="Route extends RouteId">
	import { page } from "$app/state";
	import { registerViewRail } from "$lib/app-shell.svelte";
	import type { ResolvedPathname } from "$app/types";
	import type { Snippet } from "svelte";

	type Props = {
		route: Route;
		entries: ViewRailEntry<Route>[];
		label: string;
		snippet?: Snippet;
	};
	const { route, entries, label, snippet }: Props = $props();

	registerViewRail();

	const resolvePath = resolve as unknown as (
		route: Route,
		params: ViewRailEntry<Route>["params"]
	) => ResolvedPathname;

	const paths = $derived(entries.map((entry) => resolvePath(route, entry.params)));
	const activeIndex = $derived.by(() => {
		if (page.route.id !== route) return undefined;
		return paths.findIndex((path) => page.url.pathname === path);
	});
	const ActiveComponent = $derived(
		activeIndex !== undefined && activeIndex >= 0 ? entries[activeIndex].component : undefined
	);
</script>

<div class="flex h-full flex-1 min-h-0 min-w-0 overflow-hidden bg-muted">
	<aside
		class="flex w-56 shrink-0 min-h-0 min-w-0 flex-col overflow-hidden border-e border-border bg-background"
	>
		<nav class="flex shrink-0 flex-col gap-1 p-2" aria-label={label}>
			{#each entries as entry, index (entry.label)}
				<a
					href={resolvePath(route, entry.params)}
					aria-current={index === activeIndex ? "page" : undefined}
					data-active={index === activeIndex ? "true" : undefined}
					class="flex min-w-0 items-center gap-2 rounded px-3 py-2 text-muted-foreground outline-none hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring data-[active=true]:bg-muted data-[active=true]:text-foreground"
				>
					{#if entry.icon}
						<entry.icon class="size-5 shrink-0" aria-hidden="true" />
					{/if}
					<span class="truncate">{entry.label}</span>
				</a>
			{/each}
		</nav>
		{#if snippet}
			<div class="min-h-0 flex-1 overflow-y-auto border-t border-border p-3">
				{@render snippet()}
			</div>
		{/if}
	</aside>
	<section class="flex min-w-0 min-h-0 flex-1 flex-col overflow-hidden">
		{#if ActiveComponent}
			<ActiveComponent />
		{/if}
	</section>
</div>
