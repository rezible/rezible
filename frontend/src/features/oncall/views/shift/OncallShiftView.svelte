<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftOptions, type OncallShift } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import OncallShiftDetails from "./OncallShiftDetails.svelte";
	import { setPageBreadcrumbs } from "$lib/appState.svelte";
	import { formatShiftDates } from "../../lib/shift";

	type Props = {
		shiftId: string;
	};
	const { shiftId }: Props = $props();

	const query = createQuery(() => getOncallShiftOptions({ path: { id: shiftId } }));
	const shift = $derived(query.data?.data);
	const shiftDates = $derived(shift ? formatShiftDates(shift) : "");

	setPageBreadcrumbs(() => [
		{ label: "Oncall", href: "/oncall" },
		{ label: "Shifts", href: "/oncall/shifts" },
		{ label: shiftDates, href: `/oncall/shifts/${shiftId}` },
	]);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(shift: OncallShift)}
		<OncallShiftDetails {shift} />
	{/snippet}
</LoadingQueryWrapper>