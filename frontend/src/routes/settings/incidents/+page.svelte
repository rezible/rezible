<script>
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import IncidentsSettingsView from "$features/settings/views/incidents/IncidentsSettingsView.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Incidents" }
	]);
</script>

<IncidentsSettingsView />
