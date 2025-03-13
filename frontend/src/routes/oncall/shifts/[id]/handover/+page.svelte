<script lang="ts">
	import { page } from "$app/state";
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftOptions, type OncallShift } from "$lib/api";
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import { formatShiftDates } from "$features/oncall/lib/utils";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import ShiftHandoverView from "$features/oncall/views/shift-handover/ShiftHandoverView.svelte";

	const shiftId = $derived(page.params.id);
	const shiftQuery = createQuery(() => getOncallShiftOptions({ path: { id: shiftId } }));
	const shift = $derived(shiftQuery.data?.data);
	const shiftDates = $derived(shift ? formatShiftDates(shift) : "");

	setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Shifts", href: "/oncall/shifts" },
		{ label: shiftDates, href: "/oncall/shifts/" + shiftId },
		{ label: "Handover", href: "/oncall/shifts/" + shiftId + "/handover" },
	]);
</script>

<LoadingQueryWrapper query={shiftQuery}>
	{#snippet view(shift: OncallShift)}
		<ShiftHandoverView {shift} />
	{/snippet}
</LoadingQueryWrapper>