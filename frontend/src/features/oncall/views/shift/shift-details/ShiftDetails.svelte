<script lang="ts">
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import { simulateApiDelay, type OncallShift } from "$lib/api";
	import { makeFakeComparisonMetrics, makeFakeShiftMetrics, type ComparisonMetrics, type ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { buildShiftTimeDetails, formatShiftDates, type ShiftEvent } from "$src/features/oncall/lib/utils";

	// Components
	import EventStatistics from "./EventStatistics.svelte";
	import IncidentResponseMetrics from "./IncidentResponseMetrics.svelte";
	import WorkloadBreakdown from "./WorkloadBreakdown.svelte";
	import InterruptBreakdown from "./InterruptBreakdown.svelte";
	import HealthIndicators from "./HealthIndicators.svelte";

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

	const formattedShiftDates = $derived(formatShiftDates(shift));
</script>

<div class="space-y-2">
	<WorkloadBreakdown {shift} {metrics} />

	<div class="grid grid-cols-1 md:grid-cols-2 gap-2">
		<EventStatistics {metrics} {comparison} />

		<IncidentResponseMetrics {metrics} {comparison} />
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 gap-2">
		<InterruptBreakdown {metrics} {comparison} />

		<HealthIndicators {metrics} />
	</div>
</div>
