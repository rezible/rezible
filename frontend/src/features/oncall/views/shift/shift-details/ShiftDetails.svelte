<script lang="ts">
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import { simulateApiDelay, type OncallShift } from "$lib/api";
	import { makeFakeComparisonMetrics, makeFakeShiftMetrics, type ComparisonMetrics, type ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { buildShiftTimeDetails, formatShiftDates, type ShiftEvent } from "$features/oncall/lib/utils";

	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";

	// Components
	import EventStatistics from "./EventStatistics.svelte";
	import IncidentResponseMetrics from "./IncidentResponseMetrics.svelte";
	import WorkloadBreakdown from "./WorkloadBreakdown.svelte";
	import InterruptBreakdown from "./InterruptBreakdown.svelte";

	type Props = {
		shift: OncallShift;
		shiftEvents: ShiftEvent[];
	};
	let { shift, shiftEvents }: Props = $props();

	const shiftTimeDetails = buildShiftTimeDetails(shift);

	const shiftMetricsQuery = createQuery(() =>
		queryOptions({
			queryKey: ["shiftMetrics", shift.id],
			queryFn: async (): Promise<{data: ShiftMetrics}> => {
				await simulateApiDelay(500);
				return {data: makeFakeShiftMetrics()};
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);
	const metrics = $derived(shiftMetricsQuery.data?.data);

	const shiftComparisonQuery = createQuery(() =>
		queryOptions({
			queryKey: ["shiftComparison", shift.id],
			queryFn: async (): Promise<{data: ComparisonMetrics}> => {
				await simulateApiDelay(500);
				return {data: makeFakeComparisonMetrics()}
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);
	const comparison = $derived(shiftComparisonQuery.data?.data);
</script>

<div class="w-full h-full">
	{#if metrics && comparison}
		<div class="p-2 h-full w-full overflow-y-auto space-y-2 ">
			<EventStatistics {shift} {shiftEvents} {metrics} {comparison} />

			<WorkloadBreakdown {shift} {shiftEvents} {metrics} />

			<IncidentResponseMetrics {metrics} {comparison} />

			<InterruptBreakdown {metrics} {comparison} />
		</div>
	{:else}
		<div class="grid w-full h-full place-items-center">
			<LoadingIndicator />
		</div>
	{/if}
</div>
