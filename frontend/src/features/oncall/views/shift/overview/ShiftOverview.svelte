<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftMetricsOptions } from "$lib/api";
	import { shiftViewStateCtx } from "../context.svelte";

	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";

	import ShiftEvents from "./sections/ShiftEvents.svelte";
	import ShiftIncidents from "./sections/ShiftIncidents.svelte";
	import ShiftBurden from "./sections/ShiftBurden.svelte";
	import ShiftAlerts from "./sections/ShiftAlerts.svelte";

	import ShiftEventsList from "./ShiftEventsList.svelte";

	const viewState = shiftViewStateCtx.get();
	const shiftId = $derived(viewState.shiftId);

	const comparisonQuery = createQuery(() => getOncallShiftMetricsOptions());
	const comparison = $derived(comparisonQuery.data?.data);
	
	const metricsQuery = createQuery(() => getOncallShiftMetricsOptions({query: {shiftId}}));
	const metrics = $derived(metricsQuery.data?.data);
</script>

<div class="w-full h-full grid grid-cols-3 gap-2">
	<div class="col-span-2 h-full w-full overflow-y-auto pr-1 space-y-2">
		{#if !metrics || !comparison}
			<div class="grid w-full h-full place-items-center">
				<LoadingIndicator />
			</div>
		{:else}
			<ShiftEvents {metrics} {comparison} />
			<ShiftBurden {metrics} />
			<ShiftAlerts {metrics} />
			<ShiftIncidents {metrics} {comparison} />
		{/if}
	</div>

	<div class="h-full flex flex-col overflow-y-auto">
		<ShiftEventsList />
	</div>
</div>
