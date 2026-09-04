<script lang="ts">
	import type { Incident } from "$lib/api";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initIncidentsListViewController } from "./controller.svelte";

	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/layout/filter-page/FilterPage.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { Button } from "$components/ui/button";
	import Header from "$components/layout/header/Header.svelte";

	import IncidentCard from "./IncidentCard.svelte";
	import IncidentsListViewFilters from "./IncidentsListViewFilters.svelte";
	import PageActions from "./PageActions.svelte";

	const controller = initIncidentsListViewController();

	registerPageDescriptor(() => ({
		title: "Incidents",
		actions: { component: PageActions },
	}));
</script>

<FilterPage>
	{#snippet header()}
		<Header title="Filters">
			{#snippet subheading()}
				<span class="text-xs text-muted-foreground uppercase"
					>{controller.activeFilterCount} active</span
				>
			{/snippet}

			{#snippet actions()}
				<Button
					variant="ghost"
					size="sm"
					onclick={() => {
						controller.resetFilters();
					}}
					disabled={controller.activeFilterCount === 0}
				>
					Clear Filters
				</Button>
			{/snippet}
		</Header>
	{/snippet}

	{#snippet filters()}
		<IncidentsListViewFilters />
	{/snippet}

	<PaginatedQueryListBox {...controller.paginatedIncidentsQuery}>
		<LoadingQueryWrapper query={controller.incidentsQuery}>
			{#snippet view(incidents: Incident[])}
				{#each incidents as incident (incident.id)}
					<IncidentCard {incident} />
				{:else}
					<div class="grid place-items-center min-h-48 rounded-lg border border-dashed">
						<div class="flex flex-col items-center gap-2 text-center">
							<span class="text-sm font-medium text-surface-content/80">
								No incidents found
							</span>
							{#if controller.activeFilterCount > 0}
								<Button variant="outline" size="sm" onclick={controller.resetFilters}>
									Clear filters
								</Button>
							{/if}
						</div>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
