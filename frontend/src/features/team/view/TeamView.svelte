<script lang="ts">
	import { appShell, type PageBreadcrumb } from "$features/app-shell/lib/appShellState.svelte";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { useTeamViewState } from "$features/team";
	import TeamOverview from "./overview/TeamOverview.svelte";
	import TeamBacklogView from "./backlog/TeamBacklogView.svelte";

	const view = useTeamViewState();

	const avatar = $derived<PageBreadcrumb["avatar"]>(view.team ? { kind: "team", id: view.team.id } : undefined);
	appShell.setPageBreadcrumbs(() => [
		{ label: "Teams", href: "/teams" },
		{ label: view.teamName, href: `/teams/${view.teamSlug}`, avatar },
	]);

	const tabs = $derived([
		{label: "Overview", path: "", component: TeamOverview},
		{label: "Backlog", path: "backlog", component: TeamBacklogView},
	]);
</script>

<TabbedViewContainer {tabs} pathBase="/teams/{view.teamSlug}" />
