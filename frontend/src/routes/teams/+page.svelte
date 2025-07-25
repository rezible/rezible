<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import TeamsListView from "$src/features/teams-list/view/TeamsListView.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Teams" }]);
</script>

<TeamsListView />
