<script lang="ts">
	import UserHomeView from "$features/home/views/user-home/UserHomeView.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Home" }]);
</script>

<UserHomeView />
