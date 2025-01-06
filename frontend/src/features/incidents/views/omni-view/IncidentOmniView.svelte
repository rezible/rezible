<script lang="ts">
	import { onMount } from "svelte";
	import { watch } from "runed";
    import type { IncidentViewRouteParam } from "$src/params/incidentView";
	import { type Incident, type Retrospective } from "$lib/api";

    import IncidentOverview from "$features/incidents/components/incident-overview/IncidentOverview.svelte";
    import IncidentTimeline from "$features/incidents/components/incident-timeline/IncidentTimeline.svelte";
	import IncidentAnalysis from "$features/incidents/components/incident-analysis/IncidentAnalysis.svelte";
    import IncidentFindingsReport from "$features/incidents/components/incident-report/IncidentFindingsReport.svelte";
    import ContextSidebar from "$features/incidents/components/context-sidebar/ContextSidebar.svelte";

	import { incidentCtx } from '$features/incidents/lib/context.ts';
	import { collaborationState } from '$features/incidents/lib/collaboration.svelte';
    import NavigationMenu from "./NavigationMenu.svelte";

	type Props = {
        incident: Incident;
		retrospective: Retrospective;
		viewParam: IncidentViewRouteParam;
    }
    const { incident, retrospective, viewParam }: Props = $props();

	const retroType = $derived(retrospective.attributes.type);

	incidentCtx.set(incident);

	const retrospectiveId = $derived(retrospective.id);
	watch(() => retrospectiveId, (id: string) => {collaborationState.connect(id)});
	onMount(() => {return () => {collaborationState.cleanup()}});
</script>

<div class="flex-1 min-h-0 flex flex-row gap-2 overflow-y-hidden">
	<NavigationMenu incidentSlug={incident.attributes.slug} {retroType} {viewParam} />

	<div class="flex-1 min-h-0 overflow-y-auto border p-2">
		{#if viewParam === undefined}
			<IncidentOverview {incident} />
		{:else if viewParam === "timeline"}
			<IncidentTimeline />
		{:else if viewParam === "analysis"}
			<IncidentAnalysis {incident} />
		{:else if viewParam === "report"}
			<IncidentFindingsReport {incident} {retrospective} />
		{/if}
	</div>

	<ContextSidebar />
</div>
