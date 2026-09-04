<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import TabbedViewContainer from "$src/components/layout/tabbed-view-container/TabbedViewContainer.svelte";
	import TeamOverview from "./overview/TeamOverview.svelte";
	import TeamBacklogView from "./backlog/TeamBacklogView.svelte";
	import TeamMeetings from "./meetings/TeamMeetings.svelte";
	import { initTeamViewController } from "./controller.svelte";

	const { slug }: { slug: string } = $props();

	const view = initTeamViewController(() => slug);

	registerPageDescriptor(() => ({
		title: view.teamName || "Team",
		parents: [{ label: "Teams", path: resolve("/teams") }],
	}));
</script>

<TabbedViewContainer
	route="/teams/[slug]/[[view=teamView]]"
	tabs={[
		{ label: "Overview", component: TeamOverview, params: { slug } },
		{ label: "Backlog", component: TeamBacklogView, params: { slug, view: "backlog" } },
		{ label: "Meetings", component: TeamMeetings, params: { slug, view: "meetings" } },
	]}
/>
