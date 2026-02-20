<script lang="ts">
	import { page } from "$app/state";
	import { cn } from '$lib/utils';
	import { mdiAccountGroup, mdiCog, mdiFire, mdiPuzzle } from "@mdi/js";
	import FilterPage from "$components/filter-page/FilterPage.svelte";

	const { children } = $props();

	const routes = [
		{ label: "General", route: "", icon: mdiCog },
		{ label: "Organization", route: "/organization", icon: mdiAccountGroup },
		{ label: "Incidents", route: "/incidents", icon: mdiFire },
		{ label: "Integrations", route: "/integrations", icon: mdiPuzzle },
	];
	const baseRoute = "/settings";

	const activeRoute = $derived(page.route.id?.replace(baseRoute, "") ?? "");
	const activeRouteIdx = $derived(
		routes.findIndex((p) => (p.route === "" ? activeRoute === "" : activeRoute.startsWith(p.route)))
	);
</script>

{#snippet filters()}
	<ul class="flex flex-col space-y-0 -mb-px w-full h-fit w-54 border rounded">
		{#each routes as p, i}
			{@const active = i === activeRouteIdx}
			<li class="group flex w-full" role="presentation">
				<a
					href="{baseRoute}{p.route}"
					class={cn(
						"w-full py-3 px-6 gap-2 text-lg text-center border-r-2",
						active
							? "text-primary-content border-primary bg-surface-100 active"
							: "text-surface-content border-transparent hover:text-primary-content hover:border-primary/50"
					)}
				>
					<span class="flex items-center gap-2">
						{#if p.icon}
							<!-- <Icon data={p.icon} size={24} /> -->
						{/if}
						{p.label}
					</span>
				</a>
			</li>
		{/each}
	</ul>
{/snippet}


<div class="flex gap-3 w-full h-full max-h-full pt-2 pl-2">
	<div class="w-full h-full max-h-full overflow-y-auto max-w-md flex flex-col">
		<div class="flex flex-col gap-1 border p-2 rounded overflow-x-hidden overflow-y-auto">
			{@render filters()}
		</div>
	</div>

	<div class="flex-1 block min-h-0 max-h-full overflow-y-auto">
		{@render children()}
	</div>
</div>
