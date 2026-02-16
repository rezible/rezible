<script lang="ts">
	import type { SystemComponentViewParam } from "$src/params/systemComponentView";
	import type { IdProps } from "$lib/utils.svelte";
	import { appShell } from "$features/app";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import SystemComponentOverview from "./overview/SystemComponentOverview.svelte";
	import SystemComponentIncidents from "./incidents/SystemComponentIncidents.svelte";
	import { initSystemComponentViewController } from "./controller.svelte";

	const { id }: IdProps = $props();
	const view = initSystemComponentViewController(() => id);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Components", href: "/components" },
		{ label: view.componentName, href: `/components/${view.componentId}` },
	]);

	const tabs: Tab<SystemComponentViewParam>[] = [
		{ label: "Overview", view: undefined, component: SystemComponentOverview },
		{ label: "Incidents", view: "incidents", component: SystemComponentIncidents },
	];
</script>

<TabbedViewContainer {tabs} path="/components/{view.componentId}" />