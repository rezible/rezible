<script lang="ts">
	import { useAppShell } from "$lib/app-shell.svelte";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";

	import { initOncallRosterViewController } from "./controller.svelte";
	import PageActions from "./PageActions.svelte";
	import RosterDetailsBar from "./RosterDetailsBar.svelte";

	import RosterOverview from "./overview/RosterOverview.svelte";
	import RosterMembers from "./members/RosterMembers.svelte";
	import RosterSchedule from "./schedule/RosterSchedule.svelte";
	import RosterResources from "./resources/RosterResources.svelte";
	import type { OncallRosterViewRouteParam } from "$src/params/oncallRosterView";

	const { slug }: { slug: string } = $props();

	const view = initOncallRosterViewController(() => slug);
	const rosterId = $derived(view.rosterId);

	const appShell = useAppShell();
	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Rosters", href: "/rosters" },
		{
			label: view.rosterName,
			href: `/rosters/${slug}`,
			avatar: rosterId ? {kind: "roster", id: rosterId} : undefined,
		},
	]);
	appShell.setPageActions(PageActions, true);

	const tabs: Tab<OncallRosterViewRouteParam>[] = [
		{ label: "Overview", view: undefined, component: RosterOverview },
		{ label: "Members", view: "members", component: RosterMembers },
		{ label: "Schedule", view: "schedule", component: RosterSchedule },
		{ label: "Resources", view: "resources", component: RosterResources },
	];
</script>

{#snippet infoBar()}
	<RosterDetailsBar {view} />
{/snippet}

<TabbedViewContainer {tabs} path="/rosters/{view.rosterSlug}" {infoBar} />
