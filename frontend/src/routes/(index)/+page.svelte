<script lang="ts">
	import UserHomeView from "$features/home/views/user-home/UserHomeView.svelte";
	import { setPageBreadcrumbs } from "$lib/appState.svelte";

	setPageBreadcrumbs(() => [{ label: "Home" }]);
</script>

<UserHomeView />
