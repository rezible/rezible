<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import type { OncallRosterViewRouteParam } from "$src/params/oncallRosterView";
	import { getOncallRosterOptions } from "$lib/api";
	import { appShell } from "$features/app/lib/appShellState.svelte";

	import TabbedViewContainer, {
		type Tab,
	} from "$components/tabbed-view-container/TabbedViewContainer.svelte";

	import PageActions from "./PageActions.svelte";
	import RosterOverview from "./roster-overview/RosterOverview.svelte";
	import RosterDetails from "./roster-details/RosterDetails.svelte";
	import ActiveShiftBar from "./ActiveShiftBar.svelte";

	type Props = {
		rosterId: string;
		view: OncallRosterViewRouteParam;
	};
	const { rosterId, view }: Props = $props();

	const query = createQuery(() => getOncallRosterOptions({ path: { id: rosterId } }));
	const roster = $derived(query.data?.data);
	const rosterName = $derived(roster?.attributes.name ?? "");

	appShell.setPageActions(PageActions, true);
	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Rosters", href: "/oncall/rosters" },
		{ label: rosterName, href: `/oncall/rosters/${rosterId}`, avatar: { kind: "roster", id: rosterId } },
	]);

	const tabs: Tab[] = $derived([
		{ label: "Overview", path: "" },
		{ label: "Members", path: "members" },
		{ label: "Shifts", path: "shifts" },
	]);
</script>

<TabbedViewContainer {tabs} pathBase="/oncall/rosters/{rosterId}">
	{#snippet actionsBar()}
		<ActiveShiftBar />
	{/snippet}

	{#snippet content()}
		{#if view === "members"}
			<RosterDetails />
		{:else if view === "shifts"}
			
		{:else}
			<RosterOverview />
		{/if}
	{/snippet}
</TabbedViewContainer>
