<script lang="ts">
	import { Button } from "$components/ui/button";
	import SectionHeading from "$components/common/section-heading/SectionHeading.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ListSkeleton from "$components/layout/list-skeleton/ListSkeleton.svelte";
	import IncidentRow from "$features/incidents/components/incident-row/IncidentRow.svelte";
	import { useHomeController } from "./controller.svelte";

	const controller = useHomeController();
	const incidents = $derived(controller.incidents);
	const total = $derived(controller.incidentsTotal);
</script>

{#snippet headingActions()}
	{#if total !== undefined && total > incidents.length}
		<Button variant="link" size="inline" href="/incidents?status=active">View all {total}</Button>
	{/if}
{/snippet}

<section aria-labelledby="home-incidents" class="flex min-w-0 flex-col gap-3">
	<SectionHeading id="home-incidents" title="Active incidents" count={total} actions={headingActions} />
	<LoadingQueryWrapper
		query={controller.incidentsQuery}
		feedback="quiet"
		isEmpty={(items) => items.length === 0}
	>
		{#snippet loading()}
			<ListSkeleton label="Loading active incidents" />
		{/snippet}
		{#snippet error()}
			<span>Could not load active incidents.</span>
		{/snippet}
		{#snippet empty()}
			<p class="rounded-lg border bg-card p-6 text-sm text-muted-foreground">No active incidents.</p>
		{/snippet}
		{#snippet view()}
			<ul class="divide-y overflow-hidden rounded-lg border bg-card">
				{#each incidents as incident (incident.id)}
					<IncidentRow {incident} />
				{/each}
			</ul>
		{/snippet}
	</LoadingQueryWrapper>
</section>
