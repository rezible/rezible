<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import TeamsListView from "$features/teams/views/teams-list/TeamsListView.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Teams" }]);
</script>

<TeamsListView />
