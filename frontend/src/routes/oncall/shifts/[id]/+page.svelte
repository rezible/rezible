<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { page } from "$app/state";
	import { getOncallShiftOptions, type OncallShift } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import ShiftView from "$features/oncall/views/shift/ShiftView.svelte";

	const shiftId = $derived(page.params.id);
	const query = createQuery(() =>
		getOncallShiftOptions({ path: { id: shiftId } })
	);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(shift: OncallShift)}
		<ShiftView {shift} />
	{/snippet}
</LoadingQueryWrapper>
