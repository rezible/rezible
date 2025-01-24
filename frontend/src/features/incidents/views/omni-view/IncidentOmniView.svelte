<script lang="ts">
    import type { IncidentViewRouteParam } from "$src/params/incidentView";
	import { type Incident, type Retrospective } from "$lib/api";

    import IncidentOverview from "$features/incidents/components/incident-overview/IncidentOverview.svelte";
    import IncidentRetrospective from "$src/features/incidents/components/retrospective/Retrospective.svelte";

	import { incidentCtx, retrospectiveCtx } from '$features/incidents/lib/context.ts';

	type Props = {
        incident: Incident;
		retrospective: Retrospective;
		viewParam: IncidentViewRouteParam;
    }
    const { incident, retrospective, viewParam }: Props = $props();

	incidentCtx.set(incident);
	retrospectiveCtx.set(retrospective);

	const currRoute = $derived(viewParam || "");
</script>

<div class="flex-1 min-h-0 flex flex-row gap-2 overflow-y-hidden">
	<div class="w-40 h-fit border p-2 bg-surface-200 flex flex-col gap-2 overflow-y-auto">
		{#snippet navMenuItem(label: string, route: string)}
			{@const active = (route === currRoute)}
			<a href="/incidents/{incident.attributes.slug}/{route}">
				<div class="p-2 rounded border" class:border-r-4={active} class:bg-primary-600={active} class:text-primary-content={active}>
					<span>{label}</span>
				</div>
			</a>
		{/snippet}

		<div class="flex flex-col gap-1">
			<span class="text-surface-content/75">Details</span>
			{@render navMenuItem("Overview", "")}
		</div>
		<div class="flex flex-col gap-1">
			<span class="text-surface-content/75">Retrospective</span>
			<!-- todo: check if analysis enabled -->
			{@render navMenuItem("System Analysis", "analysis")}
			{@render navMenuItem("Report", "report")}
		</div>
	</div>

	{#if viewParam === undefined}
		<div class="flex-1 min-h-0 overflow-y-auto border p-2">
			<IncidentOverview />
		</div>
	{:else}
		<IncidentRetrospective view={viewParam} />
	{/if}
</div>
