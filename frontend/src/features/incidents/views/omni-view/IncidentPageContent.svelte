<script lang="ts">
	import { incidentCtx, retrospectiveCtx } from "$features/incidents/lib/context.ts";
	import { collaboration } from "$features/incidents/lib/collaboration.svelte";
	import type { Incident, Retrospective } from "$lib/api";
	import type { IncidentViewRouteParam } from "$src/params/incidentView";

	import IncidentOverview from "$features/incidents/components/incident-overview/IncidentOverview.svelte";
	import IncidentAnalysis from "$features/incidents/components/incident-analysis/IncidentAnalysis.svelte";
	import IncidentReport from "$features/incidents/components/incident-report/IncidentReport.svelte";
	import ContextSidebar from "$features/incidents/components/context-sidebar/ContextSidebar.svelte";

	type Props = {
		incident: Incident;
		retrospective: Retrospective;
		viewParam: IncidentViewRouteParam;
	};
	const { incident, retrospective, viewParam }: Props = $props();

	incidentCtx.set(incident);
	retrospectiveCtx.set(retrospective);
	collaboration.setup();
</script>


<div class="flex-1 min-h-0 overflow-y-auto border p-2">
	{#if viewParam === undefined}
		<IncidentOverview />
	{:else if viewParam === "analysis"}
		<IncidentAnalysis />
	{:else if viewParam === "report"}
		<IncidentReport />
	{/if}
</div>

<ContextSidebar />