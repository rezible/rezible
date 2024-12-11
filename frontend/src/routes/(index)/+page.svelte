<script lang="ts">
    import PageContainer from "$components/page-container/PageContainer.svelte";
    import UserHomeView from "$features/home/views/user-home/UserHomeView.svelte";

	const pageTitle = "Good Morning, Tex";
</script>

<PageContainer breadcrumbs={[{label: pageTitle}]}>
	<UserHomeView />
</PageContainer>