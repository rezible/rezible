<script lang="ts">
	import { useSystemComponentViewState } from "$features/system-component";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import SystemComponentOverview from "./overview/SystemComponentOverview.svelte";
	import SystemComponentIncidents from "./incidents/SystemComponentIncidents.svelte";

	const view = useSystemComponentViewState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Components", href: "/components" },
		{ label: view.componentName, href: `/components/${view.componentId}` },
	]);
</script>

<TabbedViewContainer
	pathBase="/components/{view.componentId}"
	tabs={[
		{ label: "Overview", path: "", component: SystemComponentOverview },
		{ label: "Incidents", path: "incidents", component: SystemComponentIncidents },
	]}
/>