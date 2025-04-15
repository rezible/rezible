<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftMetricsOptions } from "$lib/api";
	import { shiftIdCtx } from "$features/oncall/lib/context.svelte";
	import { shiftEventMatchesFilter, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import { shiftState } from "$features/oncall/views/shift/shift.svelte";

	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";

	import EventsList from "$features/oncall/components/events-list/EventsList.svelte";
	import ShiftEvents from "./ShiftEvents.svelte";
	import IncidentMetrics from "./IncidentMetrics.svelte";
	import WorkloadBreakdown from "./WorkloadBreakdown.svelte";

	const shiftId = shiftIdCtx.get();

	const comparisonQuery = createQuery(() => getOncallShiftMetricsOptions());
	const comparison = $derived(comparisonQuery.data?.data);
	
	const metricsQuery = createQuery(() => getOncallShiftMetricsOptions({query: {shiftId}}));
	const metrics = $derived(metricsQuery.data?.data);

	let eventsFilter = $state<ShiftEventFilterKind>();
	const shiftEvents = $derived.by(() => {
		if (!eventsFilter) return shiftState.shiftEvents;
		return shiftState.shiftEvents.filter(
			(e) => !eventsFilter || shiftEventMatchesFilter(e, eventsFilter)
		);
	});
</script>

<div class="w-full h-full grid grid-cols-3 gap-2">
	<div class="col-span-2 h-full w-full overflow-y-auto pr-1 space-y-2">
		{#if !metrics || !comparison}
			<div class="grid w-full h-full place-items-center">
				<LoadingIndicator />
			</div>
		{:else}
			<ShiftEvents {shiftEvents} {metrics} {comparison} bind:eventsFilter />
			<WorkloadBreakdown {shiftEvents} {metrics} />
			<IncidentMetrics {metrics} {comparison} />
		{/if}
	</div>

	<div class="h-full flex flex-col overflow-y-auto">
		<EventsList events={shiftEvents} />
	</div>
</div>
