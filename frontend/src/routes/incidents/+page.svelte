<script lang="ts">
	import IncidentsList from "$features/incidents/views/incidents-list/IncidentsList.svelte";
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";

	setPageBreadcrumbs(() => [{ label: "Incidents" }]);
</script>

<IncidentsList />