<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { type Incident } from '$lib/api/index.js';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';
	import IncidentOverview from '$features/incidents/views/overview/IncidentOverview.svelte';
    import PageContainer, { type Breadcrumb } from '$components/page-container/PageContainer.svelte';

	const { data } = $props();

	const query = createQuery(() => data.queryOptions);
	const invalidateQuery = $derived(() => data.queryClient.invalidateQueries(data.queryOptions));

	const breadcrumbs = $derived<Breadcrumb[]>([
		{label: "Incidents", href: "/incidents"},
		{label: query.data?.data.attributes.title ?? ""},
	]);
</script>

<PageContainer {breadcrumbs}>
	<LoadingQueryWrapper {query}>
		{#snippet view(incident: Incident)}
			<IncidentOverview {incident} {invalidateQuery} />
		{/snippet}
	</LoadingQueryWrapper>
</PageContainer>