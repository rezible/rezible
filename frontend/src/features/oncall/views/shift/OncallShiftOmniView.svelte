<script lang="ts">
	import { cls } from "@layerstack/tailwind";
	import { appShell, setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import { formatShiftDates } from "$features/oncall/lib/utils";
	import { shiftIdCtx } from "$features/oncall/lib/context.svelte";
	import type { OncallShiftViewRouteParam } from "$src/params/oncallShiftView";
	import { shiftState } from "$features/oncall/lib/shift.svelte";
	import PageActions from "./PageActions.svelte";
	import ShiftDetailsBar from "./ShiftDetailsBar.svelte";

	import ShiftOverview from "./overview/ShiftOverview.svelte";
	import ShiftHandover from "./handover/ShiftHandover.svelte";

	type Props = { 
		shiftId: string;
		view?: OncallShiftViewRouteParam;
	};
	const { shiftId, view }: Props = $props();

	appShell.setPageActions(PageActions, true);
	shiftIdCtx.set(shiftId);
	shiftState.setup(shiftId);

	const tabs = [
		{label: "Overview", view: undefined},
		{label: "Handover", view: "handover"},
	];

	const shiftDates = $derived(shiftState.shift ? formatShiftDates(shiftState.shift) : "");
	const handoverBreadcrumbs = $derived(view === "handover" ? [{label: "Handover", href: `/oncall/shifts/${shiftId}/handover`}] : []);

	setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Shifts", href: "/oncall/shifts" },
		{ label: shiftDates, href: "/oncall/shifts/" + shiftId },
		...handoverBreadcrumbs,
	]);
</script>

<div class="flex flex-col h-full max-h-full min-h-0 overflow-hidden">
	<div class="w-full flex h-16 z-[1] justify-between">
		<div class="flex gap-2 self-end">
			{#each tabs as tab}
				{@const active = tab.view == view}
				<a href="/oncall/shifts/{shiftId}/{tab.view ?? ""}" 
					class={cls(
						"inline-flex self-end h-14 p-4 py-3 text-lg border border-b-0 rounded-t-lg relative", 
						active && "bg-surface-100 text-secondary",
					)}>
					{tab.label}
					<div class="absolute bottom-0 -mb-px left-0 w-full border-b border-surface-100" class:hidden={!active}></div>
				</a>
			{/each}
		</div>

		<ShiftDetailsBar />
	</div>

	<div class="flex-1 min-h-0 max-h-full overflow-y-auto border rounded-b-lg rounded-tr-lg p-2 bg-surface-100">
		{#if view == undefined}
			<ShiftOverview />
		{:else}
			<ShiftHandover />
		{/if}
	</div>
</div>
