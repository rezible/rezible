<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import type { Incident } from '$lib/api';

	import Retrospective from '$features/incidents/views/retrospective/Retrospective.svelte';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';

	const { data } = $props();

	const incidentQuery = createQuery(() => data.queryOptions);
</script>

<LoadingQueryWrapper query={incidentQuery}>
	{#snippet view(incident: Incident)}
		<Retrospective {incident} />
	{/snippet}
</LoadingQueryWrapper>