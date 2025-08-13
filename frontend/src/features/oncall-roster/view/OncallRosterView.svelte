<script lang="ts">
	import { appShell, type PageBreadcrumb } from "$features/app-shell/lib/appShellState.svelte";

	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";

	import { useOncallRosterViewState } from "$features/oncall-roster";
	import PageActions from "./PageActions.svelte";
	import RosterDetailsBar from "./RosterDetailsBar.svelte";

	import RosterOverview from "./overview/RosterOverview.svelte";
	import RosterMembers from "./members/RosterMembers.svelte";
	import RosterSchedule from "./schedule/RosterSchedule.svelte";
	import RosterResources from "./resources/RosterResources.svelte";
	import type { OncallRosterViewRouteParam } from "$src/params/oncallRosterView";

	const view = useOncallRosterViewState();

	const avatar = $derived<PageBreadcrumb["avatar"]>(view.rosterId ? {kind: "roster", id: view.rosterId} : undefined);
	const rosterBreadcrumb = $derived<PageBreadcrumb>({
		label: view.rosterName,
		href: `/rosters/${view.rosterSlug}`,
		avatar,
	});

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Rosters", href: "/rosters" },
		rosterBreadcrumb,
	]);
	appShell.setPageActions(PageActions, true);

	const tabs: Tab<OncallRosterViewRouteParam>[] = [
		{ label: "Overview", view: undefined, component: RosterOverview },
		{ label: "Members", view: "members", component: RosterMembers },
		{ label: "Schedule", view: "schedule", component: RosterSchedule },
		{ label: "Resources", view: "resources", component: RosterResources },
	];
</script>

<TabbedViewContainer {tabs} path="/rosters/{view.rosterSlug}" infoBar={RosterDetailsBar} />
