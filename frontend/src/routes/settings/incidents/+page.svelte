<script>
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import IncidentsSettingsView from "$features/settings/views/incidents/IncidentsSettingsView.svelte";

	setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Incidents" }
	]);
</script>

<IncidentsSettingsView />
