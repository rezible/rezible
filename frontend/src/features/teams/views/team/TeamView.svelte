<script lang="ts">
	import type { Team } from "$lib/api";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { setTeamViewState, TeamViewState } from "./viewState.svelte";
	import TeamOverview from "./overview/TeamOverview.svelte";
	import TeamBacklogView from "./backlog/TeamBacklogView.svelte";

	type Props = {
		team: Team;
	}
	let { team }: Props = $props();

	const id = $derived(team.id);
	const slug = $derived(team.attributes.slug);

	const viewState = new TeamViewState(() => id);
	setTeamViewState(viewState);

	// const slug = $derived(viewState)

	appShell.setPageBreadcrumbs(() => [
		{ label: "Teams", href: "/teams" },
		{ label: viewState.teamName, href: `/teams/${slug}`, avatar: { kind: "team", id } },
	]);

	const tabs = $derived([
		{label: "Overview", path: "", component: TeamOverview},
		{label: "Backlog", path: "backlog", component: TeamBacklogView},
	]);
</script>

<TabbedViewContainer {tabs} pathBase="/teams/{slug}" />
