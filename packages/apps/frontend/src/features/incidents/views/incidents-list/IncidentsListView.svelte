<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initIncidentsListViewController } from "./controller.svelte";
	import Filters from "$features/incidents/components/incident-filters/IncidentFilters.svelte";
	import Row from "$features/incidents/components/incident-row/IncidentRow.svelte";
	import { Button } from "$components/ui/button";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import PageActions from "./PageActions.svelte";

	const controller = initIncidentsListViewController();
	registerPageDescriptor(() => ({ title: "Incidents", pageActions: actions }));
</script>

{#snippet actions()}
	<PageActions />
{/snippet}

{#snippet filters()}
	<Filters
		filters={controller.filters}
		onchange={controller.setFilters}
		statusOptions={controller.statusOptions}
		severityOptions={controller.severityOptions}
	/>
{/snippet}

<FilterPage
	{filters}
	title="Incidents"
	count={controller.total}
	updating={controller.query.isFetching}
	activeFilters={controller.activeFilterCount}
	onclear={controller.resetFilters}
>
	<PaginatedQueryListBox {...controller.paginatedIncidentsQuery} showUpdating={false}>
		<LoadingQueryWrapper
			query={controller.query}
			feedback="quiet"
			isEmpty={(incidents) => incidents.length === 0}
		>
			{#snippet loading()}
				<ListSkeleton label="Loading incidents" />
			{/snippet}
			{#snippet empty()}
				<div
					class="flex flex-col items-start gap-2 rounded-lg border bg-card p-6 text-sm text-muted-foreground"
				>
					{#if controller.activeFilterCount > 0}
						<p>No incidents match these filters.</p>
						<Button variant="link" size="inline" onclick={controller.resetFilters}>
							Clear filters
						</Button>
					{:else}
						<p>No incidents yet.</p>
					{/if}
				</div>
			{/snippet}
			{#snippet view(incidents)}
				<ul class="divide-y overflow-hidden rounded-lg border bg-card">
					{#each incidents as incident (incident.id)}
						<Row {incident} />
					{/each}
				</ul>
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
