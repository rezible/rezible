<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { shiftIdCtx } from "$features/oncall/lib/context.svelte";
	import { shiftEventMatchesFilter, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import { shiftState } from "$features/oncall/views/shift/shift.svelte";

	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";

	import ShiftEvents from "./ShiftEvents.svelte";
	import IncidentMetrics from "./IncidentMetrics.svelte";
	import WorkloadBreakdown from "./WorkloadBreakdown.svelte";
	import ShiftEventsList from "./ShiftEventsList.svelte";
	import { getOncallShiftMetricsOptions } from "$src/lib/api";

	const shiftId = shiftIdCtx.get();

	const metricsQuery = createQuery(() => getOncallShiftMetricsOptions({query: {shiftId}}));
	const metrics = $derived(metricsQuery.data?.data);

	const comparisonQuery = createQuery(() => getOncallShiftMetricsOptions());
	const comparison = $derived(comparisonQuery.data?.data);

	let eventsFilter = $state<ShiftEventFilterKind>();
	const shiftEvents = $derived.by(() => {
		if (!eventsFilter) return shiftState.shiftEvents;
		return shiftState.shiftEvents.filter(
			(e) => !eventsFilter || shiftEventMatchesFilter(e, eventsFilter)
		);
	});
</script>

<div class="w-full h-full grid grid-cols-3 gap-2">
	{#if !metrics || !comparison}
		<div class="grid col-span-3 w-full h-full place-items-center">
			<LoadingIndicator />
		</div>
	{:else}
		<div class="col-span-2 h-full w-full overflow-y-auto space-y-2">
			<ShiftEvents {shiftEvents} {metrics} {comparison} bind:eventsFilter />
			<WorkloadBreakdown {shiftEvents} {metrics} />
			<IncidentMetrics {metrics} {comparison} />
		</div>

		<div class="h-full flex flex-col overflow-y-auto">
			<ShiftEventsList {shiftEvents} />
		</div>
	{/if}
</div>
