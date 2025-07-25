<script lang="ts">
	import type { IncidentViewRouteParam } from "$src/params/incidentView";
	import { appShell, type PageBreadcrumb } from "$features/app-shell/lib/appShellState.svelte";

	import { IncidentCollaborationState, setIncidentCollaboration } from "./collaboration.svelte";
	import { setIncidentViewState, IncidentViewState } from "./viewState.svelte";
	
	import PageActions from "./PageActions.svelte";
	import IncidentOverview from "./incident-overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./incident-analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./incident-report/IncidentReport.svelte";
	import ContextSidebar from "./context-sidebar/ContextSidebar.svelte";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import IncidentDetailsBar from "./IncidentDetailsBar.svelte";

	type Props = {
		incidentId: string;
		view?: IncidentViewRouteParam;
	};
	const { incidentId, view }: Props = $props();

	const viewState = new IncidentViewState(() => incidentId);
	setIncidentViewState(viewState);

	const collabState = new IncidentCollaborationState(() => viewState.retrospectiveId);
	setIncidentCollaboration(collabState);

	const incidentBreadcrumb = $derived<PageBreadcrumb>({
		label: viewState.incident?.attributes.title,
		href: `/incidents/${incidentId}`,
	});
	const retroBreadcrumb = $derived<PageBreadcrumb>({
		label: view === "analysis" ? "System Analysis" : "Report",
		href: `/incidents/${incidentId}/${view}`,
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

<TabbedViewContainer 
	pathBase="/incidents/{incidentId}" 
	infoBar={IncidentDetailsBar}
	{tabs}
>
	{#snippet tabSidebar()}
		<ContextSidebar />
	{/snippet}
</TabbedViewContainer>
