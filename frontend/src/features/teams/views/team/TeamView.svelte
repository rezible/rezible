<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { setTeamViewState, TeamViewState } from "./viewState.svelte";
	import TeamOverview from "./overview/TeamOverview.svelte";

	type Props = {
		id: string;
	}
	let { id }: Props = $props();

	const viewState = new TeamViewState(() => id);
	setTeamViewState(viewState);

	// appShell.setPageActions(PageActions, true);
	
	appShell.setPageBreadcrumbs(() => [
		{ label: "Teams", href: "/teams" },
		{ label: viewState.teamName, href: `/teams/${id}`, avatar: { kind: "team", id } },
	]);

	const tabs = $derived([
		{label: "Overview", path: "", component: TeamOverview},
		// {label: "Settings", path: "settings", component: TeamSettings},
	]);
</script>

<TabbedViewContainer {tabs} pathBase="/teams/{id}" />
