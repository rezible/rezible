<script lang="ts">
    import { createQuery } from '@tanstack/svelte-query';
	import { mdiAlien, mdiBallot, mdiFire } from '@mdi/js';
    import PageContainer, { type Breadcrumb, type PageTabsProps } from "$components/page-container/PageContainer.svelte";

	let { data, children } = $props();

	const query = createQuery(() => data.teamQueryOptions);

	const team = $derived(query.data?.data);
	const breadcrumbs = $derived<Breadcrumb[]>([
		{label: "Teams", href: "/teams"},
		{label: team?.attributes.name ?? "", avatar: team ? {kind: "team", id: team.id} : undefined},
	]);

	const tabs = $derived<PageTabsProps>({
		pages: [
			{label: "Details", path: "", icon: mdiAlien},
			{label: "Backlog", path: "/backlog", icon: mdiBallot},
			{label: "Incidents", path: "/incidents", icon: mdiFire},
		],
		basePath: `/teams/${data.slug}`,
		baseRouteId: "/teams/[slug]"
	});
</script>

<PageContainer {breadcrumbs} {tabs}>
	{@render children()}
</PageContainer>