<script lang="ts">
	import { Header } from "svelte-ux";
	import { formatDuration } from "date-fns";
	import { PieChart, Text } from "layerchart";
	import type { OncallShiftMetrics } from "$lib/api";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";
	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";

	type Props = {
		metrics: OncallShiftMetrics;
		comparison: OncallShiftMetrics;
	};

	let { metrics, comparison }: Props = $props();

	const colors = [
		'oklch(var(--color-danger))',
		'oklch(var(--color-warning))',
		'oklch(var(--color-success))',
		'oklch(var(--color-info))',
	];

	const incidentSeries = $derived(metrics.incidentActivity?.map((v, i) => ({key: v.incidentId, value: v.minutes, color: colors[i % colors.length]})));
	const totalMinutes = $derived(metrics.incidentActivity?.reduce((prev, val) => (prev + val.minutes), 0));
	const totalTimeFormatted = $derived(formatDuration({minutes: totalMinutes}, {zero: true}));
	
	const incidentStats = $derived<InlineStatProps[]>([
		{title: "Alert to Incident Ratio", subheading: `Alerts that became incidents`, value: metrics.alertIncidentRate,
			comparison: {value: comparison.alertIncidentRate, positive: true}
		},
		// {title: "Stat 2", subheading: `desc`, value: 0},
		// {title: "Stat 3", subheading: `desc`, value: 0},
		// {title: "Stat 4", subheading: `desc`, value: 0},
	])
</script>

<div class="flex flex-col gap-2 w-full p-2 border bg-surface-100/40 border-surface-content/10 rounded">
	<Header title="Incidents" subheading="" class="" />

	<ChartWithStats chart={incidentTimeChart} stats={incidentStats} />
</div>

{#snippet incidentTimeChart()}
	<div class="h-[250px] w-[300px] overflow-auto">
		<PieChart
			data={incidentSeries}
			series={incidentSeries}
			innerRadius={-20}
			cornerRadius={5}
			padAngle={0.02}
			renderContext="canvas"
		>
			<svelte:fragment slot="aboveMarks">
				<Text
					value={totalTimeFormatted}
					textAnchor="middle"
					verticalAnchor="middle"
					class="text-3xl"
					dy={4}
				/>
				<Text
					value="time spent in incidents"
					textAnchor="middle"
					verticalAnchor="middle"
					class="text-sm fill-surface-content/50"
					dy={26}
				/>
			</svelte:fragment>
		</PieChart>
	</div>
{/snippet}