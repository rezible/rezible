<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import ReportsListView from "$features/reports-list/view/ReportsListView.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Reports" }]);
</script>

<ReportsListView />