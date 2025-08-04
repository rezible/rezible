<script lang="ts">
	import type { ComponentProps } from "svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";

	import TabbedViewContainer from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { useOncallShiftViewState } from "$features/oncall-shift";

	import ShiftDetailsBar from "./ShiftDetailsBar.svelte";
	import ShiftOverview from "./overview/ShiftOverview.svelte";
	import ShiftHandover from "./handover/ShiftHandover.svelte";
	import PageActions from "./PageActions.svelte";

	const view = useOncallShiftViewState();

	const shiftBreadcrumb = $derived([{ label: view.shiftTitle, href: "/shifts/" + view.shiftId }]);
	// const handoverBreadcrumb = $derived(view === "handover" ? [{label: "Handover", href: `/shifts/${view.shiftId}/handover`}] : []);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Shifts", href: "/shifts" },
		...shiftBreadcrumb,
		// ...handoverBreadcrumb,
	]);
	const pageActionsProps = $derived<ComponentProps<typeof PageActions>>({
		previousId: view.previousShift?.id,
		nextId: view.nextShift?.id,
	})
	appShell.setPageActions(PageActions, true, () => pageActionsProps);
</script>

<TabbedViewContainer 
	pathBase="/shifts/{view.shiftId}" 
	infoBar={ShiftDetailsBar}
	tabs={[
		{label: "Overview", path: "", component: ShiftOverview},
		{label: "Handover", path: "handover", component: ShiftHandover},
	]}
/>