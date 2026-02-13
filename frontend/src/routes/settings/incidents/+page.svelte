<script>
	import { appShell } from "$features/app";
	import IncidentsSettingsView from "$features/settings/view/incidents/IncidentsSettingsView.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Incidents" }
	]);
</script>

<IncidentsSettingsView />
