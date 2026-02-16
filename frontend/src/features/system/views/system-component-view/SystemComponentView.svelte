<script lang="ts">
	import type { ComponentViewParam } from "$src/params/componentView";
	import { appShell } from "$features/app";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import SystemComponentOverview from "./overview/SystemComponentOverview.svelte";
	import SystemComponentIncidents from "./incidents/SystemComponentIncidents.svelte";
	import { initSystemComponentViewController } from "./controller.svelte";

	const { id }: { id: string } = $props();
	const view = initSystemComponentViewController(() => id);

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