<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { listOncallRostersOptions, type ListOncallRostersData, type OncallRoster } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import RosterCard from "$features/oncall-rosters-list/components/roster-card/RosterCard.svelte";
	import { QueryPaginatorState } from "$src/lib/paginator.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Rosters", href: "/rosters" },
	]);

	const paginator = new QueryPaginatorState();
	let searchValue = $state<string>();
	const params = $derived<ListOncallRostersData["query"]>({
		search: searchValue,
		...paginator.queryParams,
	});
	const query = createQuery(() => listOncallRostersOptions({ query: params }));
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(rosters: OncallRoster[])}
				{#each rosters as roster}
					<RosterCard {roster} />
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Rosters Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
