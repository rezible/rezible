<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
    import PageContainer, { type Breadcrumb, type PageTabsProps } from '$components/page-container/PageContainer.svelte';

	const { data, children } = $props();

	const query = createQuery(() => data.queryOptions);

	const breadcrumbs = $derived<Breadcrumb[]>([
		{label: "Incidents", href: "/incidents"},
		{label: query.data?.data.attributes.title ?? ""},
	]);

	const tabs = $derived<PageTabsProps>({
		pages: [
			{label: "Overview", path: ""},
			{label: "Retrospective", path: "/retrospective"}
		],
		basePath: `/incidents/${data.slug}`,
		baseRouteId: "/incidents/[slug]"
	});
</script>

<PageContainer {breadcrumbs} {tabs}>
	{@render children()}
</PageContainer>
