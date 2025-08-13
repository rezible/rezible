<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { useAlertViewState } from "$features/alert";
	import type { AlertViewParam } from "$src/params/alertView";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import AlertOverview from "./overview/AlertOverview.svelte";
	import AlertEvents from "./events/AlertEvents.svelte";
	import AlertPlaybooks from "./playbooks/AlertPlaybooks.svelte";

	const view = useAlertViewState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Alerts", href: "/alerts" },
		{ label: view.alertTitle, href: `/alerts/${view.alertId}` },
	]);

	const tabs: Tab<AlertViewParam>[] = [
		{ label: "Overview", view: undefined, component: AlertOverview },
		{ label: "Recent Activity", view: "events", component: AlertEvents },
		{ label: "Linked Playbooks", view: "playbooks", component: AlertPlaybooks },
	];
</script>

<TabbedViewContainer path="/alerts/{view.alertId}" {tabs} />