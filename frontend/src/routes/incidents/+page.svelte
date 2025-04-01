<script lang="ts">
	import IncidentsList from "$features/incidents/views/incidents-list/IncidentsList.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Incidents" }]);
</script>

<IncidentsList />