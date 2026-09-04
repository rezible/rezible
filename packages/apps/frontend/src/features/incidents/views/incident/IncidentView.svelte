<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
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

	registerPageDescriptor(() => ({
		title: controller.incident?.attributes.title ?? "Incident",
		parents: [{ label: "Incidents", path: resolve("/incidents") }],
		actions: { component: IncidentPageActions },
	}));
</script>

<TabbedViewContainer
	route="/incidents/[slug]/[[view=incidentView]]"
	tabs={[
		{ label: "Overview", component: IncidentOverview, params: { slug } },
		{ label: "Analysis", component: IncidentAnalysis, params: { slug, view: "analysis" } },
		{ label: "Report", component: IncidentReport, params: { slug, view: "report" } },
	]}
/>

<IncidentSidebar />
