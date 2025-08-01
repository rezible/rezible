<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { listIncidentsOptions, type ListIncidentsData, type Incident } from "$lib/api";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import IncidentCard from "$components/incident-card/IncidentCard.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Incidents" }]);

	const pagination = createPaginationStore();
	let searchValue = $state<string>();

	let allQueryParams = $state<ListIncidentsData["query"]>({});
	const allIncidentsQuery = createQuery(() => listIncidentsOptions({query: allQueryParams}));
</script>

{#snippet filters()}
	<SearchInput bind:value={searchValue} />
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox {pagination}>
		<LoadingQueryWrapper query={allIncidentsQuery}>
			{#snippet view(incidents: Incident[])}
				{#each incidents as incident (incident.id)}
					<IncidentCard {incident} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>