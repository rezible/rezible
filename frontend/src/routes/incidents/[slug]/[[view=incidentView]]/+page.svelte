<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { getIncidentOptions, type Incident } from '$lib/api';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';
	import IncidentView from '$features/incidents/views/incident/IncidentView.svelte';
    import PageContainer, { type Breadcrumb } from '$components/page-container/PageContainer.svelte';
    import { page } from '$app/state';

	const { data } = $props();

	const query = createQuery(() => getIncidentOptions({path: {id: data.slug}}));

	const viewParam = $derived(!page.params.view ? "overview" : "retrospective");

	const breadcrumbs = $derived<Breadcrumb[]>([
		{label: "Incidents", href: "/incidents"},
		{label: query.data?.data.attributes.title ?? ""},
	]);
</script>

<PageContainer {breadcrumbs}>
	<LoadingQueryWrapper {query}>
		{#snippet view(incident: Incident)}
			<IncidentView {incident} view={viewParam} />
		{/snippet}
	</LoadingQueryWrapper>
</PageContainer>