<script lang="ts">
	import type { ComponentProps } from "svelte";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";

	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { useOncallShiftViewState } from "$features/oncall-shift";

	import ShiftDetailsBar from "./ShiftDetailsBar.svelte";
	import ShiftOverview from "./overview/ShiftOverview.svelte";
	import ShiftHandover from "./handover/ShiftHandover.svelte";
	import PageActions from "./PageActions.svelte";
	import type { OncallShiftViewRouteParam } from "$src/params/oncallShiftView";

	const view = useOncallShiftViewState();

	const shiftBreadcrumb = $derived([{ label: view.shiftTitle, href: "/shifts/" + view.shiftId }]);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Shifts", href: "/shifts" },
		...shiftBreadcrumb,
	]);
	const pageActionsProps = $derived<ComponentProps<typeof PageActions>>({
		previousId: view.previousShift?.id,
		nextId: view.nextShift?.id,
	})
	appShell.setPageActions(PageActions, true, () => pageActionsProps);

	const tabs: Tab<OncallShiftViewRouteParam>[] = [
		{label: "Overview", view: undefined, component: ShiftOverview},
		{label: "Handover", view: "handover", component: ShiftHandover},
	];
</script>

<TabbedViewContainer {tabs} path="/shifts/{view.shiftId}" infoBar={ShiftDetailsBar} />