<script lang="ts">
	import type { OncallRosterViewRouteParam } from "$src/params/oncallRosterView";

	import { appShell, type PageBreadcrumb } from "$features/app/lib/appShellState.svelte";
	import PageActions from "./PageActions.svelte";
	import { rosterViewCtx, RosterViewState } from "./viewState.svelte";

	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import RosterDetailsBar from "./RosterDetailsBar.svelte";

	import RosterOverview from "./overview/RosterOverview.svelte";
	import RosterMembers from "./members/RosterMembers.svelte";
	import RosterSchedule from "./schedule/RosterSchedule.svelte";
	import RosterBacklog from "./backlog/RosterBacklog.svelte";
	import RosterResources from "./resources/RosterResources.svelte";

	type Props = {
		id: string;
		view: OncallRosterViewRouteParam;
	};
	const { id, view }: Props = $props();

	const viewState = new RosterViewState(() => id);
	rosterViewCtx.set(viewState);

	const rosterBreadcrumb = $derived<PageBreadcrumb>({
		label: viewState.rosterName ?? "",
		href: `/rosters/${id}`,
		avatar: { kind: "roster", id },
	});

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Rosters", href: "/rosters" },
		rosterBreadcrumb,
	]);
	appShell.setPageActions(PageActions, true, () => ({ viewState }));
</script>

<TabbedViewContainer
	pathBase="/rosters/{id}"
	infoBar={RosterDetailsBar}
	tabs={[
		{ label: "Overview", path: "", component: RosterOverview },
		{ label: "Members", path: "members", component: RosterMembers },
		{ label: "Schedule", path: "schedule", component: RosterSchedule },
		{ label: "Task Backlog", path: "backlog", component: RosterBacklog },
		{ label: "Resources", path: "resources", component: RosterResources },
	]}
/>
