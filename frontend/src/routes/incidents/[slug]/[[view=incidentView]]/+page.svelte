<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { page } from "$app/state";
	import { getIncidentOptions, getRetrospectiveForIncidentOptions } from "$lib/api";
	import { convertIncidentViewParam } from "$src/params/incidentView";

	import IncidentOmniView from "$features/incidents/views/omni-view/IncidentOmniView.svelte";
	import PageContainer, { type Breadcrumb } from "$components/page-container/PageContainer.svelte";

	const { data } = $props();

	const incQuery = createQuery(() => getIncidentOptions({ path: { id: data.slug } }));
	const incident = $derived(incQuery.data?.data);

	const retroQuery = createQuery(() => getRetrospectiveForIncidentOptions({ path: { id: data.slug } }));
	const retrospective = $derived(retroQuery.data?.data);

	const viewParam = $derived(convertIncidentViewParam(page.params.view));

	const breadcrumbs = $derived<Breadcrumb[]>([
		{ label: "Incidents", href: "/incidents" },
		{ label: incident?.attributes.title ?? "" },
		// ...(viewParam ? [{label: viewParam}] : [])
	]);
</script>

<PageContainer {breadcrumbs}>
	{#if incident && retrospective}
		{#key incident.id}
			<IncidentOmniView {incident} {retrospective} {viewParam} />
		{/key}
	{/if}
</PageContainer>
