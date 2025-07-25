<script>
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import IncidentsSettingsView from "$src/features/settings/view/incidents/IncidentsSettingsView.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Incidents" }
	]);
</script>

<IncidentsSettingsView />
