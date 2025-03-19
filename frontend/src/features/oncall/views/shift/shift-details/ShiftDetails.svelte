<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import type { ShiftEvent } from "$features/oncall/lib/utils";
	import { buildShiftTimeDetails, formatShiftDates } from "$features/oncall/lib/utils";
	import { useShiftMetricsQuery, useShiftComparisonQuery } from "$features/oncall/lib/shift-queries";
	import { Text } from "svelte-ux";
	
	// Components
	import EventStatistics from "./EventStatistics.svelte";
	import ResponseMetrics from "./ResponseMetrics.svelte";
	import WorkloadDistribution from "./WorkloadDistribution.svelte";
	import SeverityBreakdown from "./SeverityBreakdown.svelte";
	import HealthIndicators from "./HealthIndicators.svelte";

	type Props = {
		shift: OncallShift;
		shiftEvents: ShiftEvent[];
	};
	let { shift, shiftEvents }: Props = $props();
	
	const shiftTimeDetails = buildShiftTimeDetails(shift);
	const shiftMetricsQuery = useShiftMetricsQuery(shift);
	const shiftComparisonQuery = useShiftComparisonQuery(shift);
	
	$derived formattedShiftDates = formatShiftDates(shift);
	$derived isLoading = $shiftMetricsQuery.isLoading || $shiftComparisonQuery.isLoading;
</script>

<div class="space-y-6">
	<div>
		<Text variant="heading">{shift.attributes.roster.attributes.name} Shift Analysis</Text>
		<Text variant="subtitle" class="text-gray-500">{formattedShiftDates}</Text>
	</div>
	
	<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
		<EventStatistics 
			metrics={$shiftMetricsQuery.data || {
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
				burdenScore: 0
			}} 
			comparison={$shiftComparisonQuery.data || {
				alertsComparison: 0,
				incidentsComparison: 0,
				responseTimeComparison: 0,
				escalationRateComparison: 0,
				nightAlertsComparison: 0,
				severityComparison: { critical: 0, high: 0, medium: 0, low: 0 }
			}}
			loading={isLoading}
		/>
		
		<ResponseMetrics 
			metrics={$shiftMetricsQuery.data || {
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
				burdenScore: 0
			}} 
			comparison={$shiftComparisonQuery.data || {
				alertsComparison: 0,
				incidentsComparison: 0,
				responseTimeComparison: 0,
				escalationRateComparison: 0,
				nightAlertsComparison: 0,
				severityComparison: { critical: 0, high: 0, medium: 0, low: 0 }
			}}
			loading={isLoading}
		/>
	</div>
	
	<WorkloadDistribution 
		shift={shift}
		metrics={$shiftMetricsQuery.data || {
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
			burdenScore: 0
		}}
	/>
	
	<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
		<SeverityBreakdown 
			metrics={$shiftMetricsQuery.data || {
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
				burdenScore: 0
			}} 
			comparison={$shiftComparisonQuery.data || {
				alertsComparison: 0,
				incidentsComparison: 0,
				responseTimeComparison: 0,
				escalationRateComparison: 0,
				nightAlertsComparison: 0,
				severityComparison: { critical: 0, high: 0, medium: 0, low: 0 }
			}}
			loading={isLoading}
		/>
		
		<HealthIndicators 
			metrics={$shiftMetricsQuery.data || {
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
				burdenScore: 0
			}}
			loading={isLoading}
		/>
	</div>
</div>
