<script lang="ts">
	import type { Team } from "$lib/api";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { setTeamViewState, TeamViewState } from "./viewState.svelte";
	import TeamOverview from "./overview/TeamOverview.svelte";
	import TeamDetailsBar from "./TeamDetailsBar.svelte";

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
	]);
</script>

<TabbedViewContainer {tabs} pathBase="/teams/{slug}" />
