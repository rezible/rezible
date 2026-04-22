<script lang="ts">
	import { useAppShell } from "$lib/app-shell.svelte";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";

	import { initOncallRosterViewController } from "./controller.svelte";
	import PageActions from "./PageActions.svelte";
	import RosterDetailsBar from "./RosterDetailsBar.svelte";

	import RosterOverview from "./overview/RosterOverview.svelte";
	import RosterMembers from "./members/RosterMembers.svelte";
	import RosterSchedule from "./schedule/RosterSchedule.svelte";
	import RosterResources from "./resources/RosterResources.svelte";

	const { slug }: { slug: string } = $props();

	const view = initOncallRosterViewController(() => slug);

	const appShell = useAppShell();
	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Rosters", path: "/oncall/rosters" },
		{ label: view.rosterName, path: `/oncall/rosters/${slug}` },
	]);
	appShell.setPageActions(PageActions, true);
</script>

{#snippet infoBar()}
	<RosterDetailsBar {view} />
{/snippet}

<TabbedViewContainer
	route="/oncall/rosters/[slug]/[[view=oncallRosterView]]"
	{infoBar}
	tabs={[
		{ label: "Overview", component: RosterOverview, params: { slug } },
		{ label: "Members", component: RosterMembers, params: { slug, view: "members" } },
		{ label: "Schedule", component: RosterSchedule, params: { slug, view: "schedule" } },
		{ label: "Resources", component: RosterResources, params: { slug, view: "roster" } },
	]}
/>
