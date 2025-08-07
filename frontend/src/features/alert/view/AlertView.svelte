<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { useAlertViewState } from "$features/alert";
	import AlertOverview from "./overview/AlertOverview.svelte";
	import TabbedViewContainer from "$src/components/tabbed-view-container/TabbedViewContainer.svelte";

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
	]}
/>