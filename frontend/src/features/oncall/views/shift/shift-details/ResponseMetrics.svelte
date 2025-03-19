<script lang="ts">
	import { Card } from "svelte-ux";
	import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/utils";
	import { formatComparisonValue, formatDuration, formatPercentage } from "$features/oncall/lib/utils";

	type Props = {
		metrics: ShiftMetrics;
		comparison: ComparisonMetrics;
		loading: boolean;
	};

	let { metrics, comparison, loading }: Props = $props();

	const comparisonClass = (value: number) => {
		if (value > 0) return "text-red-500";
		if (value < 0) return "text-green-500";
		return "text-gray-500";
	};

	const responseComparisonClass = (value: number) => {
		if (value > 0) return "text-red-500";
		if (value < 0) return "text-green-500";
		return "text-gray-500";
	};
</script>

<Card class="p-4">
	<div class="flex items-center justify-between mb-4">
		<span>Response Metrics</span>
		{#if loading}
			<div class="text-sm text-gray-500">Loading...</div>
		{/if}
	</div>

	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		<div class="flex flex-col">
			<span>Avg Response Time (min)</span>
			<!-- <Metric value={metrics.avgResponseTime} label=")" /> -->
			<div class={responseComparisonClass(comparison.responseTimeComparison)}>
				{formatComparisonValue(-comparison.responseTimeComparison)} from average
			</div>
		</div>

		<div class="flex flex-col">
			<span>Escalation Rate</span>
			<!-- <Metric value={formatPercentage(metrics.escalationRate)} label="" /> -->
			<div class={comparisonClass(comparison.escalationRateComparison)}>
				{formatComparisonValue(comparison.escalationRateComparison)} from average
			</div>
			<div class="text-sm text-gray-500 mt-1">Alerts that became incidents</div>
		</div>

		<div class="flex flex-col">
			<span>Total Incident Time</span>
			<!-- <Metric value={formatDuration(metrics.totalIncidentTime)} label="" /> -->
			<div class="text-sm text-gray-500 mt-1">
				Longest: {formatDuration(metrics.longestIncident)}
			</div>
		</div>
	</div>
</Card>
