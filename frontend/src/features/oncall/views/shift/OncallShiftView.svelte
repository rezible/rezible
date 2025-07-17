<script lang="ts">
	import type { OncallShiftViewRouteParam } from "$src/params/oncallShiftView";
	import { appShell } from "$features/app/lib/appShellState.svelte";

	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";

	import { setShiftViewState, ShiftViewState } from "./shiftViewState.svelte";
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

	const shiftBreadcrumb = $derived([{ label: viewState.shiftTitle, href: "/shifts/" + shiftId }]);
	const handoverBreadcrumb = $derived(view === "handover" ? [{label: "Handover", href: `/shifts/${shiftId}/handover`}] : []);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Shifts", href: "/shifts" },
		...shiftBreadcrumb,
		...handoverBreadcrumb,
	]);
</script>

<TabbedViewContainer 
	pathBase="/shifts/{shiftId}" 
	infoBar={ShiftDetailsBar}
	tabs={[
		{label: "Overview", path: "", component: ShiftOverview},
		{label: "Handover", path: "handover", component: ShiftHandover},
	]}
/>