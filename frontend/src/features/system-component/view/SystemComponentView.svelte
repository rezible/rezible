<script lang="ts">
	import type { ComponentViewParam } from "$src/params/componentView";
	import { useSystemComponentViewState } from "$features/system-component";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import SystemComponentOverview from "./overview/SystemComponentOverview.svelte";
	import SystemComponentIncidents from "./incidents/SystemComponentIncidents.svelte";

	const view = useSystemComponentViewState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Components", href: "/components" },
		{ label: view.componentName, href: `/components/${view.componentId}` },
	]);

	const tabs: Tab<ComponentViewParam>[] = [
		{ label: "Overview", view: undefined, component: SystemComponentOverview },
		{ label: "Incidents", view: "incidents", component: SystemComponentIncidents },
	];
</script>

<TabbedViewContainer {tabs} path="/components/{view.componentId}" />