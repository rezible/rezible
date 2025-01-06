<script lang="ts">
	import { onMount } from "svelte";
	import { watch } from "runed";
    import type { IncidentViewRouteParam } from "$src/params/incidentView";
	import { type Incident, type Retrospective } from "$lib/api";

    import IncidentTimeline from "$features/incidents/components/incident-timeline/IncidentTimeline.svelte";
    import IncidentOverview from "$features/incidents/components/incident-overview/IncidentOverview.svelte";
    import IncidentFindingsReport from "$features/incidents/components/incident-report/IncidentFindingsReport.svelte";
	import IncidentAnalysis from "$features/incidents/components/incident-analysis/IncidentAnalysis.svelte";
    import ContextSidebar from "$features/incidents/components/context-sidebar/ContextSidebar.svelte";

	import { collaborationState } from '$features/incidents/lib/collaboration.svelte';

	type Props = {
        incident: Incident;
		retrospective: Retrospective;
		viewParam: IncidentViewRouteParam;
    }
    const { incident, retrospective, viewParam }: Props = $props();

	const viewGroups = [
		{
			label: "Details",
			children: [
				{label: "Overview", route: ""}
			]
		},
		{
			label: "Retrospective",
			children: [
				{label: "Timeline", route: "timeline"},
				{label: "Analysis", route: "analysis"},
				{label: "Findings", route: "findings"},
			]
		},
	];
	const activeViewRoute = $derived(viewParam || "");

	const retrospectiveType = $derived(retrospective.attributes.type);

	const documentName = $derived(retrospective.attributes.documentName ?? "");
	watch(() => documentName, (name: string) => {collaborationState.connect(name)});
	onMount(() => {return () => {collaborationState.cleanup()}});
</script>

<div class="flex-1 min-h-0 flex flex-row gap-2 overflow-y-hidden">
	<div class="w-40 flex flex-col gap-2 overflow-y-auto">
		{#each viewGroups as g, i}
			<div class="border p-2 bg-surface-200 flex flex-col gap-1">
				<span class="text-surface-content/75">{g.label}</span>
				{#each g.children as v}
					{@const active = (v.route === activeViewRoute)}
					<a href="/incidents/{incident.attributes.slug}/{v.route}">
						<div class="p-2 rounded border" class:border-r-4={active} class:bg-primary-600={active} class:text-primary-content={active}>
							<span>{v.label}</span>
						</div>
					</a>
				{/each}
			</div>
		{/each}
	</div>

	<div class="flex-1 min-h-0 overflow-y-auto border p-2">
		{#if activeViewRoute === ""}
			<IncidentOverview {incident} />
		{:else if activeViewRoute === "timeline"}
			<IncidentTimeline {incident} />
		{:else if activeViewRoute === "analysis"}
			<IncidentAnalysis {incident} />
		{:else if activeViewRoute === "findings"}
			<IncidentFindingsReport {incident} {retrospective} />
		{/if}
	</div>

	<ContextSidebar />
</div>
