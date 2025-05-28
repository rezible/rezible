<script lang="ts">
	import type { OncallShiftViewRouteParam } from "$src/params/oncallShiftView";
	import { appShell } from "$features/app/lib/appShellState.svelte";

	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";

	import { setShiftViewState, ShiftViewState } from "./shiftViewState.svelte";
	import PageActions from "./PageActions.svelte";
	import ShiftDetailsBar from "./ShiftDetailsBar.svelte";
	import ShiftOverview from "./overview/ShiftOverview.svelte";
	import ShiftHandover from "./handover/ShiftHandover.svelte";

	type Props = { 
		shiftId: string;
		view?: OncallShiftViewRouteParam;
	};
	const { shiftId, view }: Props = $props();

	const viewState = new ShiftViewState(() => shiftId);
	setShiftViewState(viewState);

	const shiftBreadcrumb = $derived([{ label: viewState.shiftTitle, href: "/oncall/shifts/" + shiftId }]);
	const handoverBreadcrumb = $derived(view === "handover" ? [{label: "Handover", href: `/oncall/shifts/${shiftId}/handover`}] : []);

	appShell.setPageActions(PageActions, true);
	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Shifts", href: "/oncall/shifts" },
		...shiftBreadcrumb,
		...handoverBreadcrumb,
	]);
</script>

<TabbedViewContainer 
	pathBase="/oncall/shifts/{shiftId}" 
	infoBar={ShiftDetailsBar}
	tabs={[
		{label: "Overview", path: "", component: ShiftOverview},
		{label: "Handover", path: "handover", component: ShiftHandover},
	]}
/>