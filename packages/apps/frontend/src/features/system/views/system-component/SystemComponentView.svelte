<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import SystemComponentOverview from "./overview/SystemComponentOverview.svelte";
	import SystemComponentIncidents from "./incidents/SystemComponentIncidents.svelte";
	import { initSystemComponentViewController } from "./controller.svelte";

	const { id }: IdProp = $props();
	const view = initSystemComponentViewController(() => id);

	setPageBreadcrumbs(() => [
		{ label: "System", path: "/system" },
		{ label: view.componentName, path: `/system/${view.componentId}` },
	]);
</script>

<TabbedViewContainer 
	route="/system/[id]/[[view=systemComponentView]]" 
	tabs={[
		{ label: "Overview", component: SystemComponentOverview, params: {id} },
		{ label: "Incidents", component: SystemComponentIncidents, params: {id, view: "incidents"} },
	]} 
/>