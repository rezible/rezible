<script>
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import RostersListView from "$features/oncall/views/rosters-list/RosterList.svelte";

	setPageBreadcrumbs(() => [{ label: "Oncall", href: "/oncall" }, { label: "Rosters" }])
</script>

<RostersListView />
