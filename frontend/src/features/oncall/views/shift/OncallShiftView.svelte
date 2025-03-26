<script lang="ts">
	import type { OncallShiftViewRouteParam } from "$src/params/oncallShiftView";
	import { appShell, setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import { formatShiftDates } from "$features/oncall/lib/utils";
	import { shiftIdCtx } from "$features/oncall/lib/context.svelte";
	import { shiftState } from "$features/oncall/views/shift/shift.svelte";

	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import PageActions from "./PageActions.svelte";
	import ShiftDetailsBar from "./ShiftDetailsBar.svelte";
	import ShiftOverview from "./overview/ShiftOverview.svelte";
	import ShiftHandover from "./handover/ShiftHandover.svelte";

	type Props = { 
		shiftId: string;
		view?: OncallShiftViewRouteParam;
	};
	const { shiftId, view }: Props = $props();

	shiftIdCtx.set(shiftId);
	shiftState.setup(shiftId);

	const shiftBreadcrumb = $derived(shiftState.shift ? [{ label: formatShiftDates(shiftState.shift), href: "/oncall/shifts/" + shiftId }] : []);
	const handoverBreadcrumb = $derived(view === "handover" ? [{label: "Handover", href: `/oncall/shifts/${shiftId}/handover`}] : []);

	appShell.setPageActions(PageActions, true);
	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Shifts", href: "/oncall/shifts" },
		...shiftBreadcrumb,
		...handoverBreadcrumb,
	]);

	const tabs: Tab[] = $derived([
		{label: "Overview", path: ""},
		{label: "Handover", path: "handover"},
	]);
</script>

<TabbedViewContainer {tabs} pathBase="/oncall/shifts/{shiftId}">
	{#snippet actionsBar()}
		<ShiftDetailsBar />
	{/snippet}

	{#snippet content()}
		{#if view === "handover"}
			<ShiftHandover />
		{:else}
			<ShiftOverview />
		{/if}
	{/snippet}
</TabbedViewContainer>