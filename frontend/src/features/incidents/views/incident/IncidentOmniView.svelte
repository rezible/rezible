<script lang="ts">
    import { createQuery } from "@tanstack/svelte-query";
	import { onMount } from "svelte";
	import { watch } from "runed";
    import type { IncidentViewRouteParam } from "$src/params/incidentView";
	import { getRetrospectiveForIncidentOptions, type Incident, type Retrospective } from "$lib/api";
	import { Header } from "svelte-ux";

	import { collaborationState } from '$features/incidents/lib/collaboration.svelte';
    import IncidentTimeline from "$features/incidents/components/incident-timeline/IncidentTimeline.svelte";
    import IncidentOverview from "$features/incidents/components/incident-overview/IncidentOverview.svelte";
    import FindingsReport from "$features/incidents/components/findings-report/FindingsReport.svelte";
	import SystemsAnalysis from "$features/incidents/components/systems-analysis/IncidentSystemsAnalysis.svelte";

	type Props = {
        incident: Incident;
		retrospective: Retrospective;
		viewParam: IncidentViewRouteParam;
    }
    const { incident, retrospective, viewParam }: Props = $props();

	type ViewMenuItem = {label: string; route: string};
	const views: ViewMenuItem[] = [
		{label: "Overview", route: ""},
		{label: "Timeline", route: "timeline"},
		{label: "Analysis", route: "analysis"},
		{label: "Findings", route: "findings"},
	];
	const activeViewRoute = $derived(viewParam || "");
	const activeIdx = $derived(views.findIndex(v => (v.route === activeViewRoute)));

	const incidentId = $derived(incident.id);

	const documentName = $derived(retrospective.attributes.documentName ?? "");
	watch(() => documentName, (name: string) => {collaborationState.connect(name)});
	onMount(() => {
		console.log("mounted");
		return () => {collaborationState.cleanup()}
	});
</script>

<div class="flex-1 min-h-0 flex flex-row gap-2 overflow-y-hidden">
	<div class="w-40 border flex flex-col gap-2 overflow-y-auto">
		{#each views as v, i}
			{@const active = i === activeIdx}
			<a href="/incidents/{incident.attributes.slug}/{v.route}">
				<div class="border p-2" class:border-r-4={active}>
					<span>{v.label}</span>
				</div>
			</a>	
		{/each}
	</div>

	<div class="flex-1 min-h-0 overflow-y-auto border p-2">
		{#if activeViewRoute === ""}
			<IncidentOverview {incident} />
		{:else if activeViewRoute === "timeline"}
			<IncidentTimeline {incidentId} />
		{:else if activeViewRoute === "analysis"}
			<SystemsAnalysis />
		{:else if activeViewRoute === "findings"}
			<FindingsReport {incident} {retrospective} />
		{/if}
	</div>

	<div class="w-64 border overflow-y-auto p-2">
		<Header title="Context">
			<div slot="actions">
				<span>x</span>
			</div>
		</Header>
		right sidebar
		<span>connect: {collaborationState.status}</span>
	</div>
</div>
