<script lang="ts">
	import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { formatDuration } from "date-fns";
	import { PieChart, Text } from "layerchart";
	import ChartWithStats from "$src/components/viz/ChartWithStats.svelte";
	import InlineStat, { type InlineStatProps } from "$src/components/viz/InlineStat.svelte";
	import { Header } from "svelte-ux";

	type Props = {
		metrics: ShiftMetrics;
		comparison: ComparisonMetrics;
	};

	let { metrics, comparison }: Props = $props();

	const colors = [
		'oklch(var(--color-danger))',
		'oklch(var(--color-warning))',
		'oklch(var(--color-success))',
		'oklch(var(--color-info))',
	];

	const incidentData = $derived([
		{ key: "incident-foo", value: 65 },
		{ key: "incident-bar", value: 15 },
		{ key: "incident-baz", value: 10 },
	]);
	const totalTimeFormatted = $derived(formatDuration({minutes: metrics.totalIncidentTime}));
	const colorSeries = $derived(incidentData.map((v, i) => ({key: v.key, color: colors[i % colors.length]})));
	
	const incidentStats = $derived<InlineStatProps[]>([
		{title: "Alert to Incident Ratio", subheading: `Alerts that became incidents`, value: metrics.alertIncidentRate,
			comparison: {value: comparison.escalationRateComparison, positive: true}
		},
		{title: "Stat 2", subheading: `desc`, value: ""},
		{title: "Stat 3", subheading: `desc`, value: ""},
		{title: "Stat 4", subheading: `desc`, value: ""},
	])
</script>

<div class="flex flex-col gap-2 w-full p-2 border bg-surface-100/40 border-surface-content/10 rounded">
	<Header title="Incidents" subheading="" class="" />

	<ChartWithStats chart={incidentTimeChart} stats={incidentStats} />
</div>

{#snippet incidentTimeChart()}
	<div class="h-[250px] w-[300px] overflow-auto">
		<PieChart
			data={incidentData}
			series={colorSeries}
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