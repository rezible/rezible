<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initIncidentViewController } from "./controller.svelte";
	import { initIncidentCollaborationController } from "./collaboration.svelte";

	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	import IncidentPageActions from "./PageActions.svelte";
	import IncidentOverview from "./overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./report/IncidentReport.svelte";

	import RiArticleLine from "remixicon-svelte/icons/article-line";
	import RiBarChartLine from "remixicon-svelte/icons/bar-chart-line";
	import RiDashboardLine from "remixicon-svelte/icons/dashboard-horizontal-line";

	type Props = {
		slug: string;
	};
	const { slug }: Props = $props();

	const view = initIncidentViewController(() => slug);
	initIncidentCollaborationController(() => view.retrospectiveDocumentId);

	registerPageDescriptor(() => ({
		title: view.incident?.attributes.title ?? "Incident",
		status: view.incident?.attributes.currentStatus,
		parents: [{ label: "Incidents", path: resolve("/incidents") }],
		pageActions: actions,
	}));
</script>

{#snippet actions()}
	<IncidentPageActions controller={view} />
{/snippet}

<FeatureNavigationRail
	route="/incidents/[slug]/[[view=incidentView]]"
	label="Incident"
	entries={[
		{ 
			label: "Overview", 
			icon: RiDashboardLine, 
			component: IncidentOverview, 
			params: { slug },
		},
		{
			label: "Analysis",
			icon: RiBarChartLine,
			component: IncidentAnalysis,
			params: { slug, view: "analysis" },
		},
		{
			label: "Report",
			icon: RiArticleLine,
			component: IncidentReport,
			params: { slug, view: "report" },
		},
	]}
/>
