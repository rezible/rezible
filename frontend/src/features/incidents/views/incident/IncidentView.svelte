<script lang="ts">
	import type { IncidentViewRouteParam } from "$src/params/incidentView";
	import { appShell, type PageBreadcrumb } from "$features/app/lib/appShellState.svelte";

	import { IncidentCollaborationState, setIncidentCollaboration } from "./collaboration.svelte";
	import { setIncidentViewState, IncidentViewState } from "./viewState.svelte";
	
	import PageActions from "./PageActions.svelte";
	import IncidentNavMenu from "./IncidentNavMenu.svelte";
	import CreateRetrospectiveDialog from "./CreateRetrospectiveDialog.svelte";
	import IncidentOverview from "./incident-overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./incident-analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./incident-report/IncidentReport.svelte";
	import ContextSidebar from "./context-sidebar/ContextSidebar.svelte";

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
	const pageCrumbs = $derived<PageBreadcrumb[]>(
		isIncidentView ? [incidentBreadcrumb] : [incidentBreadcrumb, retroBreadcrumb]
	);
	appShell.setPageBreadcrumbs(() => [{ label: "Incidents", href: "/incidents" }, ...pageCrumbs]);
	appShell.setPageActions(PageActions, true);
</script>

<div class="flex-1 min-h-0 flex flex-row gap-2 overflow-y-hidden">
	<IncidentNavMenu {view} />

	<div class="flex-1 min-h-0 overflow-y-auto border p-2">
		{#if view === undefined}
			<IncidentOverview />
		{:else if view === "analysis"}
			<IncidentAnalysis />
		{:else if view === "retrospective"}
			<IncidentReport />
		{/if}
	</div>

	<ContextSidebar />
</div>

{#if viewState.retrospectiveNeedsCreating}
	<CreateRetrospectiveDialog {isIncidentView} />
{/if}
