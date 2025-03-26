<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import type { OncallRosterViewRouteParam } from "$src/params/oncallRosterView";
	import { getOncallRosterOptions } from "$lib/api";
	import { appShell, type PageBreadcrumb, setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";

	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import PageActions from "./PageActions.svelte";
	import RosterStats from "./roster-stats/RosterStats.svelte";
	import RosterDetails from "./roster-details/RosterDetails.svelte";

	type Props = { 
		rosterId: string;
		view: OncallRosterViewRouteParam;
	};
	const { rosterId, view }: Props = $props();

	appShell.setPageActions(PageActions, false);

	const query = createQuery(() => getOncallRosterOptions({ path: { id: rosterId } }));
	const rosterName = $derived(query.data?.data.attributes.name ?? "");
	const rosterBreadcrumb = $derived<PageBreadcrumb[]>([{label: rosterName, href: `/oncall/rosters/${rosterId}`, avatar: { kind: "roster", id: rosterId }}]);
	setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Rosters", href: "/oncall/rosters" },
		...rosterBreadcrumb,
	]);

	const tabs: Tab[] = [
		{key: "overview", label: "Overview", href: `/oncall/rosters/${rosterId}`},
		{key: "members", label: "Members", href: `/oncall/rosters/${rosterId}/members`},
	];
</script>

<TabbedViewContainer {tabs} activeKey={view}>
	{#snippet content()}
		<span>content</span>
	{/snippet}
</TabbedViewContainer>