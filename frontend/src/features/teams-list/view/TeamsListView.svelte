<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listTeamsOptions, type ListTeamsData, type Team } from "$lib/api";
	import { appShell } from "$features/app-shell";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import TeamCard from "$components/team-card/TeamCard.svelte";
	import { QueryPaginatorState } from "$lib/paginator.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Teams" }]);

	let searchValue = $state<string>();
	const paginator = new QueryPaginatorState();
	const params = $derived(listTeamsOptions({ query: {
		limit: paginator.limit,
		offset: paginator.offset,
		search: searchValue,
	}}));
	const query = createQuery(() => params);
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox pagination={paginator.pagination}>
		<LoadingQueryWrapper {query}>
			{#snippet view(teams: Team[])}
				{#each teams as team (team.id)}
					<TeamCard {team} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
