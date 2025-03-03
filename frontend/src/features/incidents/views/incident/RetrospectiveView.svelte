<script lang="ts">
	import { incidentCtx, retrospectiveCtx } from "$features/incidents/lib/context.ts";
	import { collaboration } from "$features/incidents/lib/collaboration.svelte";
	import type { Incident, Retrospective } from "$lib/api";

	import IncidentAnalysis from "$features/incidents/components/incident-analysis/IncidentAnalysis.svelte";
	import IncidentReport from "$features/incidents/components/incident-report/IncidentReport.svelte";

	type Props = {
		incident: Incident;
		retrospective: Retrospective;
		view: "analysis" | "retrospective";
	};
	const { incident, retrospective, view }: Props = $props();

	const analysisId = $derived(retrospective.attributes.systemAnalysisId);

	incidentCtx.set(incident);
	retrospectiveCtx.set(retrospective);
	collaboration.setup(retrospective.id);
</script>

{#if view === "analysis"}
	{#if analysisId}
		<IncidentAnalysis {analysisId} />
	{:else}
		<span>no system analysis for this retrospective</span>
	{/if}
{:else if view === "retrospective"}
	<IncidentReport />
{/if}