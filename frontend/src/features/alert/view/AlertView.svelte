<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { useAlertViewState } from "$features/alert";
	import AlertOverview from "./overview/AlertOverview.svelte";

	const viewState = useAlertViewState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Alerts" },
		{ label: viewState.alertTitle, href: `/alerts/${viewState.alertId}` },
	]);
</script>

<div class="flex-1">
	<AlertOverview />
</div>