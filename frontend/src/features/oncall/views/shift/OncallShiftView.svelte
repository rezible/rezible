<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftOptions, type OncallShift } from "$lib/api";
	import { appShell, setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import { formatShiftDates } from "$features/oncall/lib/shift";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import PageActionButtons from "./PageActionButtons.svelte";
	import ShiftDetails from "./ShiftDetails.svelte";

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

	appShell.setPageActions(PageActionButtons, false);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(shift: OncallShift)}
		<ShiftDetails {shift} />
	{/snippet}
</LoadingQueryWrapper>