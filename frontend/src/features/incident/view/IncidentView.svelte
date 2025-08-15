<script lang="ts">
	import { appShell, type PageBreadcrumb } from "$features/app-shell/lib/appShellState.svelte";
	
	import PageActions from "./PageActions.svelte";
	import IncidentOverview from "./incident-overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./incident-analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./incident-report/IncidentReport.svelte";
	import ContextSidebar from "./context-sidebar/ContextSidebar.svelte";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import IncidentDetailsBar from "./IncidentDetailsBar.svelte";
	import { useIncidentViewState } from "$features/incident";
	import type { IncidentViewRouteParam } from "$src/params/incidentView";

	const view = useIncidentViewState();

	const path = $derived(`/incidents/${view.incidentSlug}`);
	const incidentBreadcrumb = $derived<PageBreadcrumb>({
		label: view.incident?.attributes.title,
		href: path,
	});

	const viewParam = $derived(view.viewRouteParam);
	const retroBreadcrumb = $derived<PageBreadcrumb>({
		label: viewParam === "analysis" ? "System Analysis" : "Report",
		href: `${path}/${viewParam}`,
	});

	const isIncidentView = $derived(viewParam === undefined);
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

<TabbedViewContainer {tabs} {path} infoBar={IncidentDetailsBar}>
	{#snippet tabSidebar()}
		<ContextSidebar />
	{/snippet}
</TabbedViewContainer>
