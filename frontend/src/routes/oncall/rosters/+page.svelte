<script>
	import RostersListView from "$features/oncall/views/rosters-list/RosterList.svelte";
	import { setPageBreadcrumbs } from "$src/lib/appState.svelte";

	setPageBreadcrumbs(() => [{ label: "Oncall", href: "/oncall" }, { label: "Rosters" }])
</script>

<RostersListView />
