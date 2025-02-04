<script lang="ts">
	import PageContainer from "$components/page-container/PageContainer.svelte";
	import TeamsListView from "$features/teams/views/teams-list/TeamsListView.svelte";
</script>

<PageContainer breadcrumbs={[{ label: "Teams" }]}>
	<TeamsListView />
</PageContainer>
