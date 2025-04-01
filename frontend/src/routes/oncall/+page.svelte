<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import UserOncallOverview from "$features/oncall/views/user-oncall-overview/UserOncallOverview.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Oncall" }]);
</script>

<UserOncallOverview />
