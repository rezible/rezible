<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { initAlertViewController } from "./controller.svelte";
	import AlertOverview from "./overview/AlertOverview.svelte";
	import AlertEvents from "./events/AlertEvents.svelte";
	import AlertPlaybooks from "./playbooks/AlertPlaybooks.svelte";
	import AlertIncidents from "./incidents/AlertIncidents.svelte";

	const { id }: IdProp = $props();

	const view = initAlertViewController(() => id);

	setPageBreadcrumbs(() => [
		{ label: "Alerts", path: "/alerts" },
		{ label: view.alertTitle, path: `/alerts/${view.alertId}` },
	]);
</script>

<TabbedViewContainer 
	route="/alerts/[id]/[[view=alertView]]"
	tabs={[
		{ label: "Overview", component: AlertOverview, params: {id} },
		{ label: "Recent Activity",  component: AlertEvents, params: {id, view: "events"} },
		{ label: "Incidents",  component: AlertIncidents, params: {id, view: "incidents"} },
		{ label: "Linked Playbooks",  component: AlertPlaybooks, params: {id, view: "playbooks"} },
	]}
/>