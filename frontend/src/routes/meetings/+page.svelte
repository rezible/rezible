<script lang="ts">
	import PageContainer from "$components/page-container/PageContainer.svelte";
	import MeetingsOverview from "$features/meetings/views/meetings-overview/MeetingsOverview.svelte";
</script>

<PageContainer breadcrumbs={[{ label: "Meetings" }]}>
	<MeetingsOverview />
</PageContainer>
