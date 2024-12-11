<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { type Incident } from '$lib/api/index.js';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';
	import IncidentOverview from '$features/incidents/views/overview/IncidentOverview.svelte';

	const { data } = $props();

	const query = createQuery(() => data.queryOptions);
	const invalidateQuery = $derived(() => data.queryClient.invalidateQueries(data.queryOptions));
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(incident: Incident)}
		<IncidentOverview {incident} {invalidateQuery} />
	{/snippet}
</LoadingQueryWrapper>