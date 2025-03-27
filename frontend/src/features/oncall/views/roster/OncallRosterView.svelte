<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import type { OncallRosterViewRouteParam } from "$src/params/oncallRosterView";
	import { getOncallRosterOptions } from "$lib/api";

	import { appShell } from "$features/app/lib/appShellState.svelte";
	import PageActions from "./PageActions.svelte";
	import { rosterIdCtx } from "./context";

	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import RosterDetailsBar from "./RosterDetailsBar.svelte";

	import RosterOverview from "./tabs/RosterOverview.svelte";
	import RosterMembers from "./tabs/RosterMembers.svelte";
	import RosterSchedule from "./tabs/RosterSchedule.svelte";
	import RosterBacklog from "./tabs/RosterBacklog.svelte";
	import RosterMetrics from "./tabs/RosterMetrics.svelte";
	import RosterResources from "./tabs/RosterResources.svelte";

	type Props = {
		rosterId: string;
		view: OncallRosterViewRouteParam;
	};
	const { rosterId, view }: Props = $props();

	rosterIdCtx.set(rosterId);

	const query = createQuery(() => getOncallRosterOptions({ path: { id: rosterId } }));
	const roster = $derived(query.data?.data);
	const rosterName = $derived(roster?.attributes.name ?? "");

	appShell.setPageActions(PageActions, true);
	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Rosters", href: "/oncall/rosters" },
		{ label: rosterName, href: `/oncall/rosters/${rosterId}`, avatar: { kind: "roster", id: rosterId } },
	]);

	const tabs = $derived([
		{ label: "Overview", path: "", component: RosterOverview },
		{ label: "Metrics", path: "metrics", component: RosterMetrics },
		{ label: "Shifts", path: "schedule", component: RosterSchedule },
		{ label: "Members", path: "members", component: RosterMembers },
		{ label: "Backlog", path: "backlog", component: RosterBacklog },
		{ label: "Resources", path: "resources", component: RosterResources },
	]);
</script>

<TabbedViewContainer {tabs} pathBase="/oncall/rosters/{rosterId}" infoBar={RosterDetailsBar} />
