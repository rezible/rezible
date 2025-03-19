<script lang="ts">
	import { Card } from "svelte-ux";
	import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/utils";
	import { formatComparisonValue } from "$features/oncall/lib/utils";

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
</script>

<Card class="p-4">
	<div class="flex items-center justify-between mb-4">
		<span>Event Statistics</span>
		{#if loading}
			<div class="text-sm text-gray-500">Loading...</div>
		{/if}
	</div>

	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		<div class="flex flex-col">
			<span>Total Alerts</span>
			<!-- <Metric value={metrics.totalAlerts} label="Total Alerts" /> -->
			<div class={comparisonClass(comparison.alertsComparison)}>
				{formatComparisonValue(comparison.alertsComparison)} from average
			</div>
		</div>

		<div class="flex flex-col">
			<span>Incidents</span>
			<!-- <Metric value={metrics.totalIncidents} label="Incidents" /> -->
			<div class={comparisonClass(comparison.incidentsComparison)}>
				{formatComparisonValue(comparison.incidentsComparison)} from average
			</div>
		</div>

		<div class="flex flex-col">
			<span>Night Alerts</span>
			<!-- <Metric value={metrics.nightAlerts} label="Night Alerts" /> -->
			<div class={comparisonClass(comparison.nightAlertsComparison)}>
				{formatComparisonValue(comparison.nightAlertsComparison)} from average
			</div>
			<div class="text-sm text-gray-500 mt-1">Potential sleep disruptions</div>
		</div>
	</div>
</Card>
