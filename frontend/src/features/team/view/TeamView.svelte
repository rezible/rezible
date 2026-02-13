<script lang="ts">
	import { appShell, type PageBreadcrumb } from "$features/app";
	import type { TeamViewParam } from "$src/params/teamView";
	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { useTeamViewState } from "$features/team";
	import TeamOverview from "./overview/TeamOverview.svelte";
	import TeamBacklogView from "./backlog/TeamBacklogView.svelte";
	import TeamMeetings from "./meetings/TeamMeetings.svelte";

	const view = useTeamViewState();

	const avatar = $derived<PageBreadcrumb["avatar"]>(view.team ? { kind: "team", id: view.team.id } : undefined);
	appShell.setPageBreadcrumbs(() => [
		{ label: "Teams", href: "/teams" },
		{ label: view.teamName, href: `/teams/${view.teamSlug}`, avatar },
	]);

	const tabs: Tab<TeamViewParam>[] = [
		{label: "Overview", view: undefined, component: TeamOverview},
		{label: "Backlog", view: "backlog", component: TeamBacklogView},
		{label: "Meetings", view: "meetings", component: TeamMeetings},
	];
</script>

<TabbedViewContainer {tabs} path="/teams/{view.teamSlug}" />
