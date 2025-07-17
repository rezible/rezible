<script lang="ts">
	import IncidentsListView from "$features/incidents/views/incidents-list/IncidentsListView.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Incidents" }]);
</script>

<IncidentsListView />