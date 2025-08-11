<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { useAlertViewState } from "$features/alert";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import AlertOverview from "./overview/AlertOverview.svelte";
	import AlertEvents from "./events/AlertEvents.svelte";
	import AlertPlaybooks from "./playbooks/AlertPlaybooks.svelte";

	const view = useAlertViewState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Alerts", href: "/alerts" },
		{ label: view.alertTitle, href: `/alerts/${view.alertId}` },
	]);
</script>

<TabbedViewContainer
	pathBase="/alerts/{view.alertId}"
	tabs={[
		{ label: "Overview", path: "", component: AlertOverview },
		{ label: "Recent Activity", path: "events", component: AlertEvents },
		{ label: "Linked Playbooks", path: "playbooks", component: AlertPlaybooks },
	]}
/>