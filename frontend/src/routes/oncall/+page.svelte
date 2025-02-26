<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import UserOncallOverview from "$features/oncall/views/user-oncall-overview/UserOncallOverview.svelte";

	setPageBreadcrumbs(() => [{ label: "Oncall" }]);
</script>

<UserOncallOverview />
