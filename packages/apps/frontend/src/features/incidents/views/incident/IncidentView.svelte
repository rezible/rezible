<script lang="ts">
	import { appShell, type PageBreadcrumb } from "$features/app";

	import type { IncidentViewRouteParam } from "$params/incidentView";
	import { initIncidentViewController, type IncidentViewParams } from "./controller.svelte";
	import { initIncidentCollaborationController } from "./collaboration.svelte";

	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import PageActions from "./PageActions.svelte";
	import ContextSidebar from "./context-sidebar/ContextSidebar.svelte";
	import IncidentOverview from "./overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./report/IncidentReport.svelte";


	const { slug, routeParam }: { slug: string, routeParam: IncidentViewRouteParam } = $props();

	const params = $derived<IncidentViewParams>({ slug, routeParam });
	const view = initIncidentViewController(() => params);
	const collab = initIncidentCollaborationController(view);

	const path = $derived(`/incidents/${slug}`);
	const incidentBreadcrumb = $derived<PageBreadcrumb>({
		label: view.incident?.attributes.title,
		href: path,
	});

	const retroBreadcrumb = $derived<PageBreadcrumb>({
		label: routeParam === "analysis" ? "System Analysis" : "Report",
		href: `${path}/analysis`,
	});

	const isIncidentView = $derived(routeParam === undefined);
	appShell.setPageBreadcrumbs(() => [
		{ label: "Incidents", href: "/incidents" }, 
		incidentBreadcrumb,
		...(isIncidentView ? [] : [retroBreadcrumb])
	]);
	appShell.setPageActions(PageActions, true);

	const tabs: Tab<IncidentViewRouteParam>[] = [
		{label: "Overview", view: undefined, component: IncidentOverview},
		{label: "System Analysis", view: "analysis", component: IncidentAnalysis},
		{label: "Report", view: "retrospective", component: IncidentReport},
	];
</script>

{#snippet infoBar()}
	<!-- TODO -->
{/snippet}

<TabbedViewContainer {tabs} {path} {infoBar}>
	{#snippet tabSidebar()}
		<ContextSidebar {collab} />
	{/snippet}
</TabbedViewContainer>
