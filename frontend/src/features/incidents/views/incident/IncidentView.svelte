<script lang="ts">
    import { createQuery } from "@tanstack/svelte-query";
    import { getRetrospectiveForIncidentOptions, type Incident } from "$lib/api";
    import IncidentTimeline from "$features/incidents/components/incident-timeline/IncidentTimeline.svelte";
    import IncidentOverview from "$features/incidents/components/incident-overview/IncidentOverview.svelte";
    import Retrospective from "$features/incidents/components/retrospective/Retrospective.svelte";
    import DiscussionsContainer from "$features/incidents/components/discussions/DiscussionsContainer.svelte";

	type Props = {
        incident: Incident;
		view: "overview" | "retrospective";
    }
    const { incident, view }: Props = $props();

	const incidentId = $derived(incident.id);

	const retrospectiveQuery = createQuery(() => ({
		...getRetrospectiveForIncidentOptions({path: {id: incident.id}}),
	}));
	const retrospective = $derived(retrospectiveQuery.data?.data);
</script>

<div class="flex-1 min-h-0 grid grid-cols-8 gap-2">
	<div class="col-span-2">
		<IncidentTimeline {incidentId} />
	</div>

	<div class="col-span-4">
		<div class="grid grid-cols-2 h-8 border">
			<div class=""><a href="/incidents/{incident.attributes.slug}">overview</a></div>
			<div class=""><a href="/incidents/{incident.attributes.slug}/retrospective">retrospective</a></div>
		</div>
		{#if view === "overview"}
			<IncidentOverview {incident} />
		{:else if view === "retrospective"}
			<Retrospective {incidentId} />
		{/if}
	</div>

	<div class="col-span-2">
		{#if retrospective}
			<DiscussionsContainer retrospectiveId={retrospective.id} />
		{/if}
	</div>
</div>