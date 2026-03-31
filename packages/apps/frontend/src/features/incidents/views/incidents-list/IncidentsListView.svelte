<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listIncidentsOptions, type ListIncidentsData, type Incident } from "$lib/api";
	import { appShell } from "$features/app";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import IncidentCard from "$components/incident-card/IncidentCard.svelte";
	import { QueryPaginatorState } from "$lib/paginator.svelte";

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
	<PaginatedListBox>
		<LoadingQueryWrapper {query}>
			{#snippet view(incidents: Incident[])}
				{#each incidents as incident (incident.id)}
					<IncidentCard {incident} />
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Incidents Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>