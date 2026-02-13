<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listTeamsOptions, type Team } from "$lib/api";
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
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(teams: Team[])}
				{#each teams as team (team.id)}
					<TeamCard {team} />
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Teams Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
