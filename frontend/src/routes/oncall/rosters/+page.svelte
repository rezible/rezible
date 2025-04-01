<script>
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import RostersListView from "$features/oncall/views/rosters-list/RosterList.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Rosters" },
	]);
</script>

<RostersListView />
