<script lang="ts">
	import { appShell, type PageBreadcrumb } from "$features/app-shell/lib/appShellState.svelte";

	import { setIncidentCollaborationState } from "../lib/collaborationState.svelte";
	
	import PageActions from "./PageActions.svelte";
	import IncidentOverview from "./incident-overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./incident-analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./incident-report/IncidentReport.svelte";
	import ContextSidebar from "./context-sidebar/ContextSidebar.svelte";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import IncidentDetailsBar from "./IncidentDetailsBar.svelte";
	import { useIncidentViewState } from "../lib/incidentViewState.svelte";

	const viewState = useIncidentViewState();
	setIncidentCollaborationState();

	const pathBase = $derived(`/incidents/${viewState.incidentSlug}`);
	const incidentBreadcrumb = $derived<PageBreadcrumb>({
		label: viewState.incident?.attributes.title,
		href: pathBase,
	});

	const view = $derived(viewState.viewRouteParam);
	const retroBreadcrumb = $derived<PageBreadcrumb>({
		label: view === "analysis" ? "System Analysis" : "Report",
		href: `${pathBase}/${view}`,
	});

	const isIncidentView = $derived(view === undefined);
	appShell.setPageBreadcrumbs(() => [
		{ label: "Incidents", href: "/incidents" }, 
		incidentBreadcrumb,
		...(isIncidentView ? [] : [retroBreadcrumb])
	]);
	appShell.setPageActions(PageActions, true);

	const tabs: Tab[] = [
		{label: "Overview", path: "", component: IncidentOverview},
		{label: "System Analysis", path: "analysis", component: IncidentAnalysis},
		{label: "Report", path: "retrospective", component: IncidentReport},
	];
</script>

<TabbedViewContainer {pathBase} {tabs} infoBar={IncidentDetailsBar}>
	{#snippet tabSidebar()}
		<ContextSidebar />
	{/snippet}
</TabbedViewContainer>
