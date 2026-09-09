<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initIncidentsListViewController } from "./controller.svelte";
	import Filters from "$features/incidents/components/incident-filters/IncidentFilters.svelte";
	import Row from "$features/incidents/components/incident-row/IncidentRow.svelte";
	import { Button } from "$components/ui/button";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import PageActions from "./PageActions.svelte";

	const controller = initIncidentsListViewController();
	registerPageDescriptor(() => ({ title: "Incidents", actions: { component: PageActions } }));
</script>

<div class="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-4">
	<div class="flex flex-wrap items-end justify-between gap-2">
		<Filters
			filters={controller.filters}
			onchange={controller.setFilters}
			severities={controller.metadataQuery.data?.data.severities}
		/>
		<Button variant="ghost" size="sm" onclick={controller.resetFilters}>Reset filters</Button>
	</div>
	<LoadingQueryWrapper query={controller.query} feedbackOnly />
	<PaginatedQueryListBox {...controller.paginatedIncidentsQuery}>
		{#if controller.query.data}
			{#each controller.query.data.data as item (item.id)}
				<Row incident={item} />
			{:else}
				<p class="p-6 text-sm text-muted-foreground">No matching incidents.</p>
			{/each}
		{/if}
	</PaginatedQueryListBox>
</div>
