<script lang="ts">
	import IncidentsList from "$features/incidents/views/incidents-list/IncidentsList.svelte";
	import { setPageBreadcrumbs } from "$lib/appState.svelte";

	setPageBreadcrumbs(() => [{ label: "Incidents" }]);
</script>

<IncidentsList />