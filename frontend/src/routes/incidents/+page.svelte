<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import IncidentsListView from "$features/incidents-list/view/IncidentsListView.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Incidents" }]);
</script>

<IncidentsListView />