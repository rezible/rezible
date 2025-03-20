<script lang="ts">
	import { Card, Grid } from "svelte-ux";
	import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { formatDelta } from "$lib/format.svelte";
	import { PieChart, Pie, Legend, Tooltip, Chart, Svg } from "layerchart";
	import LoadingIndicator from "$src/components/loader/LoadingIndicator.svelte";

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

	const gridData = [
		{
			date: new Date("2025-03-10T16:00:00.000Z"),
			value: 54,
		},
		{
			date: new Date("2025-03-11T16:00:00.000Z"),
			value: 99,
		},
		{
			date: new Date("2025-03-12T16:00:00.000Z"),
			value: 57,
		},
		{
			date: new Date("2025-03-13T16:00:00.000Z"),
			value: 55,
		},
		{
			date: new Date("2025-03-14T16:00:00.000Z"),
			value: 75,
		},
		{
			date: new Date("2025-03-15T16:00:00.000Z"),
			value: 78,
		},
		{
			date: new Date("2025-03-16T16:00:00.000Z"),
			value: 83,
		},
		{
			date: new Date("2025-03-17T16:00:00.000Z"),
			value: 83,
		},
		{
			date: new Date("2025-03-18T16:00:00.000Z"),
			value: 93,
		},
		{
			date: new Date("2025-03-19T16:00:00.000Z"),
			value: 51,
		},
	];
</script>

<Card class="p-4">
	<div class="flex items-center justify-between mb-4">
		<span>Interrupt Breakdown</span>
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
		{#if !(metrics && comparison)}
			<LoadingIndicator />
		{:else}
			<div class="h-64">
				<div class="h-[300px] p-4 border rounded">
					<Chart
						data={gridData}
						x="date"
						y="value"
						yDomain={[0, 100]}
						padding={{ top: 20, bottom: 20, left: 20, right: 20 }}
					>
						<Svg>
							<Grid x y />
						</Svg>
					</Chart>
				</div>
			</div>
		{/if}
	</div>
</Card>
