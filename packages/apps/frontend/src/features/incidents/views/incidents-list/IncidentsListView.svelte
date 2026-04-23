<script lang="ts">
	import type { Incident } from "$lib/api";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import { useAppShell } from "$lib/app-shell.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";
	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";
	import { initIncidentsListViewController } from "./controller.svelte";
	import IncidentCard from "./IncidentCard.svelte";
	import IncidentsListViewFilters from "./IncidentsListViewFilters.svelte";
	import PageActions from "./PageActions.svelte";

	setPageBreadcrumbs(() => [{ label: "Incidents" }]);

	const controller = initIncidentsListViewController();
	const appShell = useAppShell();
	appShell.setPageActions(PageActions, true);
</script>

<FilterPage>
	{#snippet header()}
		<Header title="Filters">
			{#snippet subheading()}
				<span class="text-xs text-muted-foreground uppercase">{controller.activeFilterCount} active</span>
			{/snippet}

			{#snippet actions()}
				<Button
					variant="ghost"
					size="sm"
					onclick={controller.resetFilters}
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

	<PaginatedListBox>
		<LoadingQueryWrapper query={controller.incidentsQuery}>
			{#snippet view(_: Incident[])}
				{#each controller.incidents as incident (incident.id)}
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
	</PaginatedListBox>
</FilterPage>
