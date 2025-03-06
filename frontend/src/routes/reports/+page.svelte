<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import ReportsDashboard from "$features/reports/views/reports-dashboard/ReportsDashboard.svelte";

	setPageBreadcrumbs(() => [{ label: "Reports" }]);
</script>

<ReportsDashboard />