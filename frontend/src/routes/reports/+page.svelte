<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/appState.svelte";
	import ReportsList from "$features/reports/views/reports-list/ReportsList.svelte";

	setPageBreadcrumbs(() => [{ label: "Reports" }]);
</script>

<ReportsList />