<script lang="ts">
    import PageContainer from "$components/page-container/PageContainer.svelte";
    import UserOncallDisplay from "$features/oncall/views/user-oncall-overview/UserOncallOverview.svelte";
</script>

<PageContainer breadcrumbs={[{label: "Oncall"}]}>
	<UserOncallDisplay />
</PageContainer>