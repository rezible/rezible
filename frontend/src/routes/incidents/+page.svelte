<script lang="ts">
	import PageContainer from "$components/page-container/PageContainer.svelte";
	import IncidentsList from "$features/incidents/views/incidents-list/IncidentsList.svelte";
</script>

<PageContainer breadcrumbs={[{ label: "Incidents" }]}>
	<IncidentsList />
</PageContainer>
