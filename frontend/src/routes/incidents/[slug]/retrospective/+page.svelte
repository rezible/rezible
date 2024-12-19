<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { type Incident } from '$lib/api/index.js';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';
    import PageContainer, { type Breadcrumb, type PageTabsProps } from '$components/page-container/PageContainer.svelte';
	import Retrospective from '$features/incidents/views/retrospective/RetrospectiveView.svelte';
	const { data } = $props();

	const incidentQuery = createQuery(() => data.queryOptions);
	const incAttr = $derived(incidentQuery.data?.data.attributes)

	const breadcrumbs = $derived<Breadcrumb[]>([
		{label: "Incidents", href: "/incidents"},
		{label: incAttr?.title ?? "", href: `/incidents/${incAttr?.slug}`},
		{label: "Retrospective"}
	]);
</script>

<PageContainer {breadcrumbs}>
	<LoadingQueryWrapper query={incidentQuery}>
		{#snippet view(incident: Incident)}
			{@const incidentId = incident.id}
			{#key incidentId}
				<Retrospective {incidentId} />
			{/key}
		{/snippet}
	</LoadingQueryWrapper>
</PageContainer>