<script lang="ts">
	import { useAppShell, type PageBreadcrumb } from "$lib/app-shell.svelte";
	import type { IncidentViewRouteParam } from "$params/incidentView";
	import { initIncidentViewController } from "./controller.svelte";

	import TabbedViewContainer from "$components/layout/tabbed-view-container/TabbedViewContainer.svelte";
	import IncidentPageActions from "./PageActions.svelte";
	import IncidentSidebar from "./sidebar/IncidentSidebar.svelte";
	import IncidentOverview from "./overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./report/IncidentReport.svelte";

	type Props = {
		slug: string;
		param: IncidentViewRouteParam;
	};
	const { slug, param }: Props = $props();

	const controller = initIncidentViewController(() => slug);

	const incidentTitle = $derived(controller.incident?.attributes.title);

	const breadcrumbs = $derived<PageBreadcrumb[]>([
		{ label: "Incidents", path: "/incidents" }, 
		{ label: incidentTitle, path: `/incidents/${slug}` },
	]);

	const appShell = useAppShell();
	appShell.setPageBreadcrumbs(() => breadcrumbs);
	appShell.setPageActions(IncidentPageActions, true);
</script>

<TabbedViewContainer 
	route="/incidents/[slug]/[[view=incidentView]]"
	tabs={[
		{label: "Overview", component: IncidentOverview, params: {slug}},
		{label: "Analysis", component: IncidentAnalysis, params: {slug, view: "analysis"}},
		{label: "Report", component: IncidentReport, params: {slug, view: "report"}},
	]}
/>

<IncidentSidebar />