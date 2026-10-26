<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { incidentResponseStatus } from "$features/incidents/lib/status";
	import { initIncidentViewController } from "./controller.svelte";
	import { initIncidentCollaborationController } from "./collaboration.svelte";
	import { initIncidentResponseController } from "./response.svelte";

	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	import IncidentPageActions from "./PageActions.svelte";
	import IncidentResponseDialogs from "./IncidentResponseDialogs.svelte";
	import IncidentOverview from "./overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./report/IncidentReport.svelte";

	import RiArticleLine from "remixicon-svelte/icons/article-line";
	import RiBarChartLine from "remixicon-svelte/icons/bar-chart-line";
	import RiDashboardLine from "remixicon-svelte/icons/dashboard-horizontal-line";

	type Props = {
		ref: string;
	};
	const { ref }: Props = $props();

	const view = initIncidentViewController(() => ref);
	initIncidentCollaborationController(
		() => view.retrospectiveDocumentId,
		() => view.reviewClosed,
		view.refetchRetrospective
	);
	const response = initIncidentResponseController();

	registerPageDescriptor(() => ({
		title: view.incident?.attributes.title ?? "Incident",
		reference: view.reference,
		status: view.incident ? incidentResponseStatus(view.incident.attributes.responseState) : undefined,
		parents: [{ label: "Incidents", path: resolve("/incidents") }],
		related: view.relatedLinks,
		contentHeading: view.activeView === undefined,
		pageActions: actions,
	}));
</script>

{#snippet actions()}
	<IncidentPageActions controller={view} {response} />
{/snippet}

<LoadingQueryWrapper query={view.incidentQuery} feedbackOnly />

{#if view.incident}
	<IncidentResponseDialogs />
	<FeatureNavigationRail
		route="/incidents/[ref]/[[view=incidentView]]"
		label="Incident"
		entries={[
			{
				label: "Overview",
				icon: RiDashboardLine,
				component: IncidentOverview,
				params: { ref },
			},
			{
				label: "Analysis",
				icon: RiBarChartLine,
				component: IncidentAnalysis,
				params: { ref, view: "analysis" },
			},
			{
				label: "Report",
				icon: RiArticleLine,
				component: IncidentReport,
				params: { ref, view: "report" },
			},
		]}
	/>
{/if}
