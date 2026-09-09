<script lang="ts">
	import { useHomeController } from "./controller.svelte";
	import { Button } from "$components/ui/button";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import Filters from "$features/incidents/components/incident-filters/IncidentFilters.svelte";
	import Row from "$features/incidents/components/incident-row/IncidentRow.svelte";

	const controller = useHomeController();
	const query = $derived(controller.incidentsQuery);
</script>

<section aria-labelledby="home-incidents" class="min-w-0 rounded-lg border border-border bg-card">
	<header class="flex items-center justify-between gap-2 p-3">
		<h2 id="home-incidents" class="font-semibold">
			Incidents
			<span class="ml-2 text-sm font-normal text-muted-foreground">
				{query.data?.pagination.total ?? "—"}
			</span>
		</h2>
		<a class="text-sm text-primary hover:underline" href={controller.incidentsHref}>View all</a>
	</header>
	<div class="space-y-2 border-b border-border px-3 pb-3">
		<Filters
			filters={controller.incidents}
			onchange={controller.setIncidents}
			severities={controller.metadataQuery.data?.data.severities}
		/>
		<Button variant="ghost" size="sm" onclick={controller.resetIncidents}>Reset filters</Button>
	</div>
	{#if controller.metadataQuery.isError}
		<p class="px-3 pt-2 text-xs text-destructive">
			Severity options could not be loaded.
			<button
				class="underline"
				onclick={() => controller.metadataQuery.refetch()}
			>
				Try again
			</button>
		</p>
	{/if}
	<LoadingQueryWrapper {query} feedbackOnly />
	{#if query.data}
		<div data-preview="incidents">
			{#each query.data.data as item (item.id)}
				<Row incident={item} />
				{:else}
					<p class="p-6 text-center text-sm text-muted-foreground">No matching incidents.</p>
			{/each}
		</div>
	{/if}
</section>
