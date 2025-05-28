<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftBurdenMetricWeightsOptions, getOncallShiftMetricsOptions } from "$lib/api";
	import { shiftViewStateCtx } from "../context.svelte";

	import ShiftEventsTotalCards from "./sections/ShiftEventsTotalCards.svelte";
	import ShiftIncidents from "./sections/ShiftIncidents.svelte";
	import ShiftBurden from "./sections/ShiftBurden.svelte";
	import ShiftAlerts from "./sections/ShiftAlerts.svelte";

	import ShiftEventsList from "./ShiftEventsList.svelte";
	import ShiftEventsHeatmap from "./sections/ShiftEventsHeatmap.svelte";

	const viewState = shiftViewStateCtx.get();
	const shiftId = $derived(viewState.shiftId);

	const comparisonQuery = createQuery(() => getOncallShiftMetricsOptions());
	const comparison = $derived(comparisonQuery.data?.data);
	
	const metricsQuery = createQuery(() => getOncallShiftMetricsOptions({query: {shiftId}}));
	const metrics = $derived(metricsQuery.data?.data);

	const burdenWeightsQuery = createQuery(() => getOncallShiftBurdenMetricWeightsOptions());
	const burdenWeights = $derived(burdenWeightsQuery.data?.data);
</script>

<div class="w-full h-full grid grid-cols-3 gap-2">
	<div class="col-span-2 h-full w-full overflow-y-auto pr-1 space-y-2">
		<ShiftEventsTotalCards {metrics} {comparison} />
		<ShiftBurden {metrics} weights={burdenWeights} />
		<ShiftEventsHeatmap onDayClicked={(d) => {}} />
		<ShiftAlerts {metrics} />
		<ShiftIncidents {metrics} {comparison} />
	</div>

	<div class="h-full flex flex-col overflow-y-auto">
		<ShiftEventsList />
	</div>
</div>
