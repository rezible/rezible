<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import type { Incident } from '$lib/api';

	import Retrospective from '$features/incidents/views/retrospective/RetrospectiveView.svelte';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';

	const { data } = $props();

	const incidentQuery = createQuery(() => data.queryOptions);
</script>

<LoadingQueryWrapper query={incidentQuery}>
	{#snippet view(incident: Incident)}
		{@const incidentId = incident.id}
		{#key incidentId}
			<Retrospective {incidentId} />
		{/key}
	{/snippet}
</LoadingQueryWrapper>