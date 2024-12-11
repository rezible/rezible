<script lang="ts">
    import PageContainer from "$components/page-container/PageContainer.svelte";
    import ReportDashboardView from "$features/reports/views/dashboard/ReportDashboardView.svelte";
</script>

<PageContainer breadcrumbs={[{label: "Reports"}]}>
	<ReportDashboardView />
</PageContainer>

