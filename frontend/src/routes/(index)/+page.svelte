<script lang="ts">
	import UserHomeView from "$features/home/view/UserHomeView.svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Home" }]);
</script>

<UserHomeView />
