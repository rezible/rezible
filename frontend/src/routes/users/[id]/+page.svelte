<script lang="ts">
	import PageContainer from "$components/page-container/PageContainer.svelte";
</script>

<PageContainer breadcrumbs={[{ label: "Users", href: "/users" }, {label: "Name", href: "/users/foo"}]}>
	<span>todo user</span>
</PageContainer>
