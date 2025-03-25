<script lang="ts">
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import {
		makeFakeComparisonMetrics,
		makeFakeShiftMetrics,
		type ComparisonMetrics,
		type ShiftMetrics,
	} from "$features/oncall/lib/shift-metrics";
	import { shiftIdCtx } from "$features/oncall/lib/context.svelte";
	import { shiftEventMatchesFilter, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import { shiftState } from "$features/oncall/views/shift/shift.svelte";

	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";

	import ShiftEvents from "./ShiftEvents.svelte";
	import IncidentMetrics from "./IncidentMetrics.svelte";
	import WorkloadBreakdown from "./WorkloadBreakdown.svelte";
	import ShiftEventsList from "./ShiftEventsList.svelte";

	const shiftId = shiftIdCtx.get();

	const shiftMetricsQuery = createQuery(() =>
		queryOptions({
			queryKey: ["shiftMetrics", shiftId],
			queryFn: async (): Promise<{ data: ShiftMetrics }> => {
				return { data: makeFakeShiftMetrics() };
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);
	const metrics = $derived(shiftMetricsQuery.data?.data);

	const shiftComparisonQuery = createQuery(() =>
		queryOptions({
			queryKey: ["shiftComparison", shiftId],
			queryFn: async (): Promise<{ data: ComparisonMetrics }> => {
				return { data: makeFakeComparisonMetrics() };
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);
	const comparison = $derived(shiftComparisonQuery.data?.data);

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
