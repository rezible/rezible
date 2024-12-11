<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { getIncidentUserDebriefOptions, getRetrospectiveForIncidentOptions } from '$lib/api/index.js';
	import Retrospective from '$features/incidents/views/retrospective/Retrospective.svelte';

	const { data } = $props();

	const incidentQuery = createQuery(() => data.queryOptions);
	const incident = $derived(incidentQuery.data?.data);

	const incidentId = $derived(incident?.id ?? "");
	const debriefQuery = createQuery(() => ({
		...getIncidentUserDebriefOptions({path: {id: incidentId}}),
		enabled: !!incidentId,
	}));
	const debrief = $derived(debriefQuery.data?.data);

	const retrospectiveQuery = createQuery(() => ({
		...getRetrospectiveForIncidentOptions({path: {id: incidentId}}),
		enabled: !!incidentId,
	}));
	const retrospective = $derived(retrospectiveQuery.data?.data);
</script>

{#if incident && retrospective && debrief}
	{#key incidentId}
		<Retrospective {incident} {retrospective} {debrief} />
	{/key}
{/if}