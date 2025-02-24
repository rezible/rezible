<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import {
		getOncallShiftOptions,
		type OncallShift,
	} from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import ShiftHandoverDetails from "./ShiftHandoverDetails.svelte";
	import { setPageBreadcrumbs } from "$lib/appState.svelte";
	import { formatShiftDates } from "$features/oncall/lib/shift";

	type Props = {
		shiftId: string;
	};
	const { shiftId }: Props = $props();

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
		<ShiftHandoverDetails {shift} />
	{/snippet}
</LoadingQueryWrapper>