<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import ReportsList from "$features/reports/views/reports-list/ReportsList.svelte";

	setPageBreadcrumbs(() => [{ label: "Reports" }]);
</script>

<ReportsList />