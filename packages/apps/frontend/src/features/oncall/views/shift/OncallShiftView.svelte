<script lang="ts">
	import { useAppShell } from "$lib/app-shell.svelte";

	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { initOncallShiftViewController } from "./controller.svelte";

	import ShiftDetailsBar from "./ShiftDetailsBar.svelte";
	import ShiftOverview from "./overview/ShiftOverview.svelte";
	import ShiftHandover from "./handover/ShiftHandover.svelte";
	import PageActions from "./PageActions.svelte";

	const { id }: IdProp = $props();

	const view = initOncallShiftViewController(() => id);

	const appShell = useAppShell();
	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Shifts", path: "/oncall/shifts" },
		{ label: view.shiftTitle, path: `/oncall/shifts/${view.shiftId}` },
	]);
	appShell.setPageActions(PageActions, true);
</script>

{#snippet infoBar()}
	<ShiftDetailsBar {view} />
{/snippet}

<TabbedViewContainer 
	route="/oncall/shifts/[id]/[[view=oncallShiftView]]"
	{infoBar}
	tabs={[

		{label: "Overview", component: ShiftOverview, params: {id}},
		{label: "Handover", component: ShiftHandover, params: {id, view: "handover"}},
	]}
/>