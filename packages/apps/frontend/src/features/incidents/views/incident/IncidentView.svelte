<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initIncidentViewController } from "./controller.svelte";

	import FeatureNavigationRail from "$components/layout/feature-navigation-rail/FeatureNavigationRail.svelte";
	
	import IncidentPageActions from "./PageActions.svelte";
	import IncidentSidebar from "./sidebar/IncidentSidebar.svelte";
	import IncidentOverview from "./overview/IncidentOverview.svelte";
	import IncidentAnalysis from "./analysis/IncidentAnalysis.svelte";
	import IncidentReport from "./report/IncidentReport.svelte";

	type Props = {
		slug: string;
	};
	const { slug }: Props = $props();

	const controller = initIncidentViewController(() => slug);

	registerPageDescriptor(() => ({
		title: controller.incident?.attributes.title ?? "Incident",
		parents: [{ label: "Incidents", path: resolve("/incidents") }],
		actions: { component: IncidentPageActions },
	}));
</script>

<FeatureNavigationRail
	route="/incidents/[slug]/[[view=incidentView]]"
	label="Incident"
	entries={[
		{ label: "Overview", component: IncidentOverview, params: { slug } },
		{ label: "Analysis", component: IncidentAnalysis, params: { slug, view: "analysis" } },
		{ label: "Report", component: IncidentReport, params: { slug, view: "report" } },
	]}
/>

<IncidentSidebar />
