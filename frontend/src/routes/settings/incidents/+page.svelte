<script>
	import { setPageBreadcrumbs } from "$lib/appState.svelte";
	import IncidentsSettingsView from "$features/settings/views/incidents/IncidentsSettingsView.svelte";

	setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Incidents" }
	]);
</script>

<IncidentsSettingsView />
