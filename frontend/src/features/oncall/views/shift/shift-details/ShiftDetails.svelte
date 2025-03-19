<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import type { ComparisonMetrics, ShiftEvent, ShiftMetrics } from "$features/oncall/lib/utils";
	import { buildShiftTimeDetails, formatShiftDates } from "$features/oncall/lib/utils";

	// Components
	import EventStatistics from "./EventStatistics.svelte";
	import ResponseMetrics from "./ResponseMetrics.svelte";
	import WorkloadDistribution from "./WorkloadDistribution.svelte";
	import SeverityBreakdown from "./SeverityBreakdown.svelte";
	import HealthIndicators from "./HealthIndicators.svelte";
	import { Header } from "svelte-ux";
	import { createQuery, queryOptions } from "@tanstack/svelte-query";

	type Props = {
		shift: OncallShift;
		shiftEvents: ShiftEvent[];
	};
	let { shift, shiftEvents }: Props = $props();

	const shiftTimeDetails = buildShiftTimeDetails(shift);

	const shiftMetricsQuery = createQuery(() =>
		queryOptions({
			queryKey: ["shiftMetrics", shift.id],
			queryFn: async (): Promise<ShiftMetrics> => {
				// Simulate API delay
				await new Promise((resolve) => setTimeout(resolve, 500));

				// Return mock data
				return {
					totalAlerts: Math.floor(Math.random() * 15) + 5,
					totalIncidents: Math.floor(Math.random() * 5) + 1,
					nightAlerts: Math.floor(Math.random() * 6) + 1,
					avgResponseTime: Math.floor(Math.random() * 20) + 5,
					escalationRate: Math.floor(Math.random() * 30) + 10,
					totalIncidentTime: Math.floor(Math.random() * 240) + 60,
					longestIncident: Math.floor(Math.random() * 120) + 30,
					businessHoursAlerts: Math.floor(Math.random() * 10) + 3,
					offHoursAlerts: Math.floor(Math.random() * 8) + 2,
					peakAlertHour: Math.floor(Math.random() * 24),
					totalOncallTime: 24 * 60, // 24 hours in minutes
					severityBreakdown: {
						critical: Math.floor(Math.random() * 2),
						high: Math.floor(Math.random() * 3) + 1,
						medium: Math.floor(Math.random() * 4) + 1,
						low: Math.floor(Math.random() * 3),
					},
					sleepDisruptionScore: Math.floor(Math.random() * 70) + 10,
					workloadScore: Math.floor(Math.random() * 80) + 20,
					burdenScore: Math.floor(Math.random() * 75) + 15,
				};
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);
	const shiftComparisonQuery = createQuery(() =>
		queryOptions({
			queryKey: ["shiftComparison", shift.id],
			queryFn: async (): Promise<ComparisonMetrics> => {
				// Simulate API delay
				await new Promise((resolve) => setTimeout(resolve, 700));

				// Return mock comparison data (percentage difference from average)
				return {
					alertsComparison: Math.floor(Math.random() * 60) - 30, // -30% to +30%
					incidentsComparison: Math.floor(Math.random() * 70) - 35,
					responseTimeComparison: Math.floor(Math.random() * 50) - 25,
					escalationRateComparison: Math.floor(Math.random() * 40) - 20,
					nightAlertsComparison: Math.floor(Math.random() * 80) - 40,
					severityComparison: {
						critical: Math.floor(Math.random() * 60) - 30,
						high: Math.floor(Math.random() * 50) - 25,
						medium: Math.floor(Math.random() * 40) - 20,
						low: Math.floor(Math.random() * 30) - 15,
					},
				};
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);

	const emptyMetricsData = {
		totalAlerts: 0,
		totalIncidents: 0,
		nightAlerts: 0,
		avgResponseTime: 0,
		escalationRate: 0,
		totalIncidentTime: 0,
		longestIncident: 0,
		businessHoursAlerts: 0,
		offHoursAlerts: 0,
		peakAlertHour: 0,
		totalOncallTime: 0,
		severityBreakdown: { critical: 0, high: 0, medium: 0, low: 0 },
		sleepDisruptionScore: 0,
		workloadScore: 0,
		burdenScore: 0,
	};

	const emptyComparisonData = {
		alertsComparison: 0,
		incidentsComparison: 0,
		responseTimeComparison: 0,
		escalationRateComparison: 0,
		nightAlertsComparison: 0,
		severityComparison: { critical: 0, high: 0, medium: 0, low: 0 },
	}

	const formattedShiftDates = $derived(formatShiftDates(shift));
	const isLoading = $derived(shiftMetricsQuery.isLoading || shiftComparisonQuery.isLoading);
</script>

<div class="space-y-6">
	<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
		<EventStatistics
			metrics={shiftMetricsQuery.data || emptyMetricsData}
			comparison={shiftComparisonQuery.data || emptyComparisonData}
			loading={isLoading}
		/>

		<ResponseMetrics
			metrics={shiftMetricsQuery.data || emptyMetricsData}
			comparison={shiftComparisonQuery.data || emptyComparisonData}
			loading={isLoading}
		/>
	</div>

	<WorkloadDistribution
		{shift}
		metrics={shiftMetricsQuery.data || emptyMetricsData}
	/>

	<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
		<SeverityBreakdown
			metrics={shiftMetricsQuery.data || emptyMetricsData}
			comparison={shiftComparisonQuery.data || emptyComparisonData}
			loading={isLoading}
		/>

		<HealthIndicators
			metrics={shiftMetricsQuery.data || emptyMetricsData}
			loading={isLoading}
		/>
	</div>
</div>
