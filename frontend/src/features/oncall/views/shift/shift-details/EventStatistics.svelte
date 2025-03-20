<script lang="ts">
	import { Card } from "svelte-ux";
	import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { formatDelta } from "$lib/format.svelte";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";

	type Props = {
		metrics?: ShiftMetrics;
		comparison?: ComparisonMetrics;
	};

	let { metrics, comparison }: Props = $props();

	const comparisonClass = (value: number) => {
		if (value > 0) return "text-red-500";
		if (value < 0) return "text-green-500";
		return "text-gray-500";
	};
</script>

<Card class="p-4">
	<div class="flex items-center justify-between mb-4">
		<span>Event Statistics</span>
	</div>

	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		{#if !(metrics && comparison)}
			<LoadingIndicator />
		{:else}
			<div class="flex flex-col">
				<span>Total Alerts</span>
				<span class="text-lg font-bold">{metrics.totalAlerts}</span>
				<div class={comparisonClass(comparison.alertsComparison)}>
					{formatDelta(comparison.alertsComparison)} from average
				</div>
			</div>

			<div class="flex flex-col">
				<span>Incidents</span>
				<span class="text-lg font-bold">{metrics.totalIncidents}</span>
				<div class={comparisonClass(comparison.incidentsComparison)}>
					{formatDelta(comparison.incidentsComparison)} from average
				</div>
			</div>

			<div class="flex flex-col">
				<span>Night Alerts</span>
				<span class="text-lg font-bold">{metrics.nightAlerts}</span>
				<div class={comparisonClass(comparison.nightAlertsComparison)}>
					{formatDelta(comparison.nightAlertsComparison)} from average
				</div>
				<div class="text-sm text-gray-500 mt-1">Potential sleep disruptions</div>
			</div>
		{/if}
	</div>
</Card>
