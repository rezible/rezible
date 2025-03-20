<script lang="ts">
	import { BarStack, Card, Tooltip } from "svelte-ux";
	import { formatDelta } from "$lib/format.svelte";
	import { settings } from "$lib/settings.svelte";
	import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import { cls } from "@layerstack/tailwind";
	import { formatDuration } from "date-fns";

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

	const responseComparisonClass = (value: number) => {
		if (value > 0) return "text-red-500";
		if (value < 0) return "text-green-500";
		return "text-gray-500";
	};

	const incidentBarData = $derived([
		{ label: 'incident-foo', value: 65, classes: { bar: 'bg-warning' } },
		{ label: 'incident-bar', value: 15, classes: { bar: 'bg-info' } },
		{ label: 'incident-baz', value: 10, classes: { bar: 'bg-success' } },
	])
	
</script>

<Card class="p-4">
	<div class="flex items-center justify-between mb-4">
		<span>Incident Response Metrics</span>
	</div>

	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		{#if !(metrics && comparison)}
			<LoadingIndicator />
		{:else}
			<div class="flex flex-col">
				<span>Alert-to-Incident Rate</span>
				<span class="text-lg font-bold">{metrics.alertIncidentRate}%</span>
				<div class={comparisonClass(comparison.escalationRateComparison)}>
					{formatDelta(comparison.escalationRateComparison)} from average
				</div>
				<div class="text-sm text-gray-500 mt-1">Alerts that became incidents</div>
			</div>

			<div class="flex flex-col">
				<span>Total Incident Time</span>
				<span class="text-lg font-bold mb-2">{formatDuration({minutes: metrics.totalIncidentTime})}</span>
				<BarStack data={incidentBarData} let:item let:total>
					<Tooltip
					  title="{item.label}: {settings.format(item.value / total, 'percent')} ({formatDuration({minutes: item.value})})"
					  placement="bottom-start"
					  offset={2}
					>
					  <div
						class={cls(
						  "h-1 group-first:rounded-l group-last:rounded-r",
						  item.classes?.bar,
						)}
					  ></div>
					  <div class="truncate text-xs font-semibold text-surface-content">
						{item.label}
					  </div>
					</Tooltip>
				</BarStack>
				  
			</div>
		{/if}
	</div>
</Card>
