<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { listOncallRostersOptions, type ListOncallRostersData, type OncallRoster } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import RosterCard from "$features/oncall-rosters-list/components/roster-card/RosterCard.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Rosters", href: "/rosters" },
	]);

	const pagination = createPaginationStore();

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
	<PaginatedListBox {pagination}>
		<LoadingQueryWrapper query={allQuery}>
			{#snippet view(rosters: OncallRoster[])}
				{#each rosters as roster}
					<RosterCard {roster} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
