<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import TabbedViewContainer from "$src/components/layout/tabbed-view-container/TabbedViewContainer.svelte";
	import { initAlertViewController } from "./controller.svelte";
	import AlertOverview from "./overview/AlertOverview.svelte";
	import AlertEvents from "./events/AlertEvents.svelte";
	import AlertIncidents from "./incidents/AlertIncidents.svelte";

	const { id }: IdProp = $props();

	const view = initAlertViewController(() => id);

	registerPageDescriptor(() => ({
		title: view.alertTitle || "Signal",
		parents: [{ label: "Signals", path: resolve("/signals") }],
	}));
</script>

<TabbedViewContainer
	route="/signals/[id]/[[view=signalView]]"
	tabs={[
		{ label: "Overview", component: AlertOverview, params: { id } },
		{ label: "Recent Activity", component: AlertEvents, params: { id, view: "events" } },
		{ label: "Incidents", component: AlertIncidents, params: { id, view: "incidents" } },
	]}
/>
