<script lang="ts">
	import type { OncallShiftBurdenMetricWeights, OncallShiftMetrics, OncallShiftMetricsBurden } from "$lib/api";

	import Header from "$components/header/Header.svelte";
	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";

	import * as echarts from "echarts";
	import EChart, { type ChartProps } from "$components/viz/echart/EChart.svelte";
	import SectionCard from "./SectionCard.svelte";

	const defaultBurden: OncallShiftMetricsBurden = {
		finalScore: 0,
		interruption: 0,
		lifeImpact: 0,
		responseRequirements: 0,
		support: 0,
		timeImpact: 0
	};

	type Props = {
		metrics?: OncallShiftMetrics;
		weights?: OncallShiftBurdenMetricWeights;
	};
	const { metrics, weights }: Props = $props();

	const burden = $derived(!!metrics ? metrics.burden : defaultBurden);

	const getScoreLabel = (score: number) => {
		if (score < 30) return "Low";
		if (score < 70) return "Moderate";
		return "High";
	};

	const burdenStatCategories = $derived<InlineStatProps[]>([
		{
			title: "Event Frequency",
			subheading: `description`,
			value: 0,
		},
		{
			title: "Life Impact",
			subheading: `description`,
			value: 0,
		},
		{
			title: "Response Requirements",
			subheading: `description`,
			value: 0,
		},
		{
			title: "Time Impact",
			subheading: `description`,
			value: 0,
		},
		{
			title: "Support",
			subheading: `description`,
			value: 0,
		},
	]);

	const burdenGaugeValues = [10, 10, 10, 10, 10];

	const burdenGaugeData: echarts.GaugeSeriesOption["data"] = [
		{
			name: "Burden\nScore",
			value: burdenGaugeValues.reduce((prev, curr) => (prev + curr)),
			title: {
				offsetCenter: ["0%", "-25%"],
			},
			detail: {
				offsetCenter: ["0%", "-5%"],
			},
			progress: {
				show: false,
				width: 0,
			}
		},
		{
			name: "Event\nFrequency",
			value: 10,
			title: {
				offsetCenter: ["-120%", "50%"],
			},
			detail: {
				offsetCenter: ["-120%", "70%"],
			},
		},
		{
			name: "Life\nImpact",
			value: 20,
			title: {
				offsetCenter: ["-60%", "40%"],
			},
			detail: {
				offsetCenter: ["-60%", "60%"],
			},
		},
		{
			name: "Response\nRequirements",
			value: 30,
			title: {
				offsetCenter: ["0%", "50%"],
			},
			detail: {
				offsetCenter: ["0%", "70%"],
			},
		},
		{
			name: "Time\nImpact",
			value: 40,
			title: {
				offsetCenter: ["60%", "40%"],
			},
			detail: {
				offsetCenter: ["60%", "60%"],
				formatter: "10",
			},
		},
		{
			name: "Support",
			value: 50,
			title: {
				offsetCenter: ["110%", "55%"],
			},
			detail: {
				offsetCenter: ["110%", "70%"],
			},
		},
	];

	const burdenChartOptions = $derived<ChartProps["options"]>({
		series: [
			{
				type: "gauge",
				name: "Burden Score",
				data: burdenGaugeData,
				startAngle: 180,
      			endAngle: 0,
				min: 0,
				max: 100,
				pointer: {
					show: false,
				},
				progress: {
					show: true,
					clip: true,
					width: 18,
					itemStyle: {
						borderWidth: 0,
					}
				},
				axisLine: {
					show: true,
					lineStyle: {
						shadowBlur: 0,
						opacity: .10,
						width: 18,
					}
				},
				tooltip: {
					formatter: (p) => {
						const val = burdenGaugeValues.at(p.dataIndex - 1);
						return `${p.name}: ${val}`;
					}
				},
				axisLabel: { color: "inherit" },
				title: {
					fontSize: 14,
					color: "inherit",
					show: true,
				},
				detail: {
					width: 30,
					height: 10,
					fontSize: 14,
					borderRadius: 3,
					backgroundColor: "inherit",
					color: "black",
					formatter: (val: number) => `${val}`
				},
			},
		],
		tooltip: {
			formatter: "{b} : {c}%",
		},
	});
</script>

<SectionCard>
	<Header title="Shift Burden" subheading="Indicator of the human impact of this shift" />
	<ChartWithStats stats={burdenStatCategories}>
		{#snippet chart()}
			<div class="h-[350px] w-[400px] overflow-hidden grid place-self-center">
				<EChart init={echarts.init} options={burdenChartOptions} />
			</div>
		{/snippet}
	</ChartWithStats>
</SectionCard>
