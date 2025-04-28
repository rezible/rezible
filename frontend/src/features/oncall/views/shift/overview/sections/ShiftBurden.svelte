<script lang="ts">
	import type { OncallShiftMetrics } from "$lib/api";

	import { formatDuration } from "date-fns";
	import { Header } from "svelte-ux";

	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";

	import * as echarts from "echarts";
	import EChart, { type ChartProps } from "$components/viz/echart/EChart.svelte";

	type Props = {
		metrics: OncallShiftMetrics;
	};
	let { metrics }: Props = $props();

	const getScoreLabel = (score: number) => {
		if (score < 30) return "Low";
		if (score < 70) return "Moderate";
		return "High";
	};

	const burdenStats = $derived<InlineStatProps[]>([
		{
			title: "High Severity Incidents",
			subheading: "Incidents with a severity of 1 or 2",
			value: metrics.incidents,
		},
		{
			title: "Sleep Disruption",
			subheading: `Based on ${metrics.nightAlerts} night alerts`,
			value: metrics.nightAlerts,
		},
		{
			title: "KTLO Workload",
			subheading: `Based on backlog and ongoing incidents`,
			value: getScoreLabel(metrics.workloadScore),
		},
		{
			title: "Off-Hours Activity",
			subheading: `Time spent active outside of 8am-6pm`,
			value: formatDuration({ minutes: metrics.offHoursActivityTime }, { zero: true }),
		},
	]);

	const burdenGaugeData = [
		{
			value: 20,
			name: "Metric 1",
			title: {
				offsetCenter: ["-75%", "85%"],
			},
			detail: {
				offsetCenter: ["-75%", "105%"],
			},
		},
		{
			value: 40,
			name: "Metric 2",
			title: {
				offsetCenter: ["0%", "85%"],
			},
			detail: {
				offsetCenter: ["0%", "105%"],
			},
		},
		{
			value: 35,
			name: "Metric 3",
			title: {
				offsetCenter: ["75%", "85%"],
			},
			detail: {
				offsetCenter: ["75%", "105%"],
			},
		},
	];

	const burdenChartOptions = $derived<ChartProps["options"]>({
		tooltip: {
			formatter: "{b} : {c}%",
		},
		series: [
			{
				type: "gauge",
				name: "Burden Score",
				anchor: { show: false },
				pointer: {
					icon: "path://M2.9,0.7L2.9,0.7c1.4,0,2.6,1.2,2.6,2.6v115c0,1.4-1.2,2.6-2.6,2.6l0,0c-1.4,0-2.6-1.2-2.6-2.6V3.3C0.3,1.9,1.4,0.7,2.9,0.7z",
					width: 3,
					length: "55%",
				},
				progress: { show: true, overlap: false, width: 6 },
				axisLine: { show: false },
				axisLabel: { color: "inherit" },
				data: burdenGaugeData,
				title: {
					fontSize: 14,
					color: "inherit",
				},
				detail: {
					width: 30,
					height: 10,
					fontSize: 12,
					backgroundColor: "inherit",
					borderRadius: 3,
					formatter: "{value}%",
				},
			},
		],
	});
</script>

<div class="flex flex-col gap-2 w-full p-2 border border-surface-content/10 rounded">
	<Header title="Burden Rating" subheading="Indicator of the human impact of this shift" />
	<ChartWithStats stats={burdenStats}>
		{#snippet chart()}
			<div class="h-[300px] w-[300px] overflow-hidden grid place-self-center">
				<EChart init={echarts.init} options={burdenChartOptions} />
			</div>
		{/snippet}
	</ChartWithStats>
</div>
