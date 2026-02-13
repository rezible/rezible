<script lang="ts">
	import { page } from "$app/state";
	import { cn } from '$lib/utils';
	import { mdiAccount, mdiAccountGroup, mdiCog, mdiFire, mdiPuzzle } from "@mdi/js";
	import FilterPage from "$components/filter-page/FilterPage.svelte";

	const { children } = $props();

	const pages = [
		{ label: "General", path: "", routeId: "/(general)", icon: mdiCog },
		{ label: "Incidents", path: "/incidents", icon: mdiFire },
		{ label: "Integrations", path: "/integrations", icon: mdiPuzzle },
		{ label: "Users", path: "/users", icon: mdiAccount },
		{ label: "Teams", path: "/teams", icon: mdiAccountGroup },
	];
	const basePath = "/settings";

	const activePath = $derived(page.route.id?.replace("/settings", "") ?? "");
	const activePageIdx = $derived(pages.findIndex((p) => (p.routeId ?? p.path) === activePath));
</script>

{#snippet filters()}
	<ul class="flex flex-col space-y-0 -mb-px w-full h-fit w-54 border rounded">
		{#each pages as p, i}
			{@const active = i === activePageIdx}
			<li class="group flex w-full" role="presentation">
				<a
					href="{basePath}{p.path}"
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

<FilterPage {filters}>
	{@render children()}
</FilterPage>

