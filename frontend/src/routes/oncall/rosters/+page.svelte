<script>
    import PageContainer from "$components/page-container/PageContainer.svelte";
	import RostersListView from "$features/oncall/views/rosters-list/RosterList.svelte";
</script>

<PageContainer breadcrumbs={[{label: "Oncall", href: "/oncall"}, {label: "Rosters"}]}>
	<RostersListView />
</PageContainer>