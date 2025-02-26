<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import TeamsListView from "$features/teams/views/teams-list/TeamsListView.svelte";

	setPageBreadcrumbs(() => [{ label: "Teams" }]);
</script>

<TeamsListView />
