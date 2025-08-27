<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listIncidentsOptions, type ListIncidentsData, type Incident } from "$lib/api";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import IncidentCard from "$components/incident-card/IncidentCard.svelte";
	import { QueryPaginatorState } from "$src/lib/paginator.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Incidents" }]);

	const paginator = new QueryPaginatorState();
	let searchValue = $state<string>();
	const params = $derived<ListIncidentsData["query"]>({
		search: searchValue,
		...paginator.queryParams,
	});
	const query = createQuery(() => listIncidentsOptions({ query: params }));
	paginator.watchQuery(query);
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox pagination={paginator.pagination}>
		<LoadingQueryWrapper {query}>
			{#snippet view(incidents: Incident[])}
				{#each incidents as incident (incident.id)}
					<IncidentCard {incident} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>