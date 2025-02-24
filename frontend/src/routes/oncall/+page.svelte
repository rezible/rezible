<script lang="ts">
	import UserOncallOverview from "$features/oncall/views/user-oncall-overview/UserOncallOverview.svelte";
	import { setPageBreadcrumbs } from "$src/lib/appState.svelte";

	setPageBreadcrumbs(() => [{ label: "Oncall" }]);
</script>

<UserOncallOverview />
