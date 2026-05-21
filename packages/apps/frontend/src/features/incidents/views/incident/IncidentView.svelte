<script lang="ts">
	import { useAppShell, type PageBreadcrumb } from "$lib/app-shell.svelte";
	import type { IncidentViewRouteParam } from "$params/incidentView";
	import { initIncidentViewController } from "./controller.svelte";

	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import PageActions from "./PageActions.svelte";
	import ContextSidebar from "./context-sidebar/ContextSidebar.svelte";
	import IncidentOverview from "./overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./report/IncidentReport.svelte";

	type Props = {
		slug: string;
		routeParam: IncidentViewRouteParam;
	};
	const { slug, routeParam }: Props = $props();
	const view = initIncidentViewController(() => slug);

	const incidentTitle = $derived(view.incident?.attributes.title);

	const breadcrumbs = $derived<PageBreadcrumb[]>([
		{ label: "Incidents", path: "/incidents" }, 
		{ label: incidentTitle, path: `/incidents/${slug}` },
	]);

	const appShell = useAppShell();
	appShell.setPageBreadcrumbs(() => breadcrumbs);
	appShell.setPageActions(PageActions, true);
</script>

<TabbedViewContainer 
	route="/incidents/[slug]/[[view=incidentView]]"
	tabs={[
		{label: "Overview", component: IncidentOverview, params: {slug}},
		{label: "System Analysis", component: IncidentAnalysis, params: {slug, view: "analysis"}},
		{label: "Report", component: IncidentReport, params: {slug, view: "report"}},
	]}
/>

<ContextSidebar collab={view.collab} />