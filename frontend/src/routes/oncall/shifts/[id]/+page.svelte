<script lang="ts">
	import { page } from "$app/state";
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftOptions, type OncallShift } from "$lib/api";
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import { formatShiftDates } from "$src/features/oncall/lib/utils";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import OncallShiftView from "$features/oncall/views/shift/OncallShiftView.svelte";

	const shiftId = $derived(page.params.id);
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
		<OncallShiftView {shift} />
	{/snippet}
</LoadingQueryWrapper>
