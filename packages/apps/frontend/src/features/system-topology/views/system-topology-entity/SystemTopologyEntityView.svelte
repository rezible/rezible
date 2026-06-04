<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import TabbedViewContainer from "$src/components/layout/tabbed-view-container/TabbedViewContainer.svelte";
	import SystemTopologyEntityOverview from "./overview/SystemTopologyEntityOverview.svelte";
	import SystemTopologyEntityIncidents from "./incidents/SystemTopologyEntityIncidents.svelte";
	import { initSystemTopologyEntityViewController } from "./controller.svelte";

	const { id }: IdProp = $props();
	const view = initSystemTopologyEntityViewController(() => id);

	setPageBreadcrumbs(() => [
		{ label: "System", path: "/system" },
		{ label: view.entityName, path: `/system/${view.entityId}` },
	]);
</script>

<TabbedViewContainer 
	route="/system/[id]/[[view=systemTopologyEntityView]]"
	tabs={[
		{ label: "Overview", component: SystemTopologyEntityOverview, params: {id} },
		{ label: "Incidents", component: SystemTopologyEntityIncidents, params: {id, view: "incidents"} },
	]} 
/>
