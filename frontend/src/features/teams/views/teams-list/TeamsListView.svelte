<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { listTeamsOptions, type ListTeamsData, type Team } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import TeamCard from "$components/team-card/TeamCard.svelte";

	const pagination = createPaginationStore();
	let searchValue = $state<string>();
	let params = $state<ListTeamsData["query"]>({});
	const query = createQuery(() => listTeamsOptions({ query: params }));

	// const maybeUpdateParams = (newParams: QueryParams) => {
	// 	let { limit, offset, search } = params || {};
	// 	offset = offset ?? 0;

	// 	const newLimit = newParams?.limit ?? limit;
	// 	const newOffset = newParams?.offset ?? offset;
	// 	const newSearch = newParams?.search ?? search;

	// 	if (limit !== newLimit || offset !== newOffset || search !== newSearch) {
	// 		// console.log('updateParams', { limit: newLimit, offset: newOffset, search: newSearch });
	// 	}
	// };
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox {pagination}>
		<LoadingQueryWrapper {query}>
			{#snippet view(teams: Team[])}
				{#each teams as team (team.id)}
					<TeamCard {team} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
