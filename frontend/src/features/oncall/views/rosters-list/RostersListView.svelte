<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { ListItem, Button } from "svelte-ux";
	import { mdiChevronRight } from "@mdi/js";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import { listOncallRostersOptions, type ListOncallRostersData, type OncallRoster } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Rosters", href: "/rosters" },
	]);

	let searchValue = $state<string>();

	let queryParams = $derived<ListOncallRostersData["query"]>({});
	const allQuery = createQuery(() => listOncallRostersOptions({query: queryParams}));

	const updateSearch = (value: any) => {
		console.log(value);
	};
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<div class="flex flex-col h-full gap-2 overflow-x-hidden overflow-y-auto">
		<LoadingQueryWrapper query={allQuery}>
			{#snippet view(rosters: OncallRoster[])}
				{#each rosters as r}
					<a href="/rosters/{r.attributes.slug}">
						<ListItem title={r.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
							<svelte:fragment slot="avatar">
								<Avatar kind="roster" size={32} id={r.id} />
							</svelte:fragment>
							<div slot="actions">
								<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
							</div>
						</ListItem>
					</a>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</FilterPage>

