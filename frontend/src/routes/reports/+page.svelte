<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import ReportsDashboard from "$features/reports/views/reports-dashboard/ReportsDashboard.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Reports" }]);
</script>

<ReportsDashboard />