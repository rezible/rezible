<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/appState.svelte";
	import TeamsListView from "$features/teams/views/teams-list/TeamsListView.svelte";

	setPageBreadcrumbs(() => [{ label: "Teams" }]);
</script>

<TeamsListView />
