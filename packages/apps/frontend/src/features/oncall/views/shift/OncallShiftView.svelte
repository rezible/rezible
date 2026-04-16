<script lang="ts">
	import { useAppShell } from "$lib/appShell.svelte";

	import TabbedViewContainer, { type Tab } from "$components/tabbed-view-container/TabbedViewContainer.svelte";
	import { initOncallShiftViewController } from "./controller.svelte";

	import ShiftDetailsBar from "./ShiftDetailsBar.svelte";
	import ShiftOverview from "./overview/ShiftOverview.svelte";
	import ShiftHandover from "./handover/ShiftHandover.svelte";
	import PageActions from "./PageActions.svelte";
	import type { OncallShiftViewRouteParam } from "$src/params/oncallShiftView";

	const { id }: IdProp = $props();

	const view = initOncallShiftViewController(() => id);

	const shiftBreadcrumb = $derived([{ label: view.shiftTitle, href: "/shifts/" + view.shiftId }]);

	const appShell = useAppShell();
	appShell.setPageBreadcrumbs(() => [
		{ label: "Oncall Shifts", href: "/shifts" },
		...shiftBreadcrumb,
	]);
	appShell.setPageActions(PageActions, true);

	const tabs: Tab<OncallShiftViewRouteParam>[] = [
		{label: "Overview", view: undefined, component: ShiftOverview},
		{label: "Handover", view: "handover", component: ShiftHandover},
	];
</script>

{#snippet infoBar()}
	<ShiftDetailsBar {view} />
{/snippet}

<TabbedViewContainer {tabs} path="/shifts/{view.shiftId}" {infoBar} />