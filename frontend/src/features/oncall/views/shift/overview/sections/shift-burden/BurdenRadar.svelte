<script lang="ts">
	import * as echarts from "echarts";
	import EChart from "$components/viz/echart/EChart.svelte";
	import type { InlineStatProps } from "$components/viz/InlineStat.svelte";

	type Props = {
		burdenValue: number;
		burdenStats: InlineStatProps[];
		comparisonSetName?: string;
	};
	const { burdenValue, burdenStats, comparisonSetName = "Roster Average" }: Props = $props();

	const indicators = burdenStats.map(v => ({
		name: v.title.replaceAll(" ", "\n"),
		min: 0,
		max: 10,
	}));

	const comparisonStats = burdenStats.map(v => (v.comparison?.value || 0));
	const shiftStats = burdenStats.map(v => (v.value || 0));

	const radarAreaSplitColors = [
		"rgb(30, 250, 30)",
		"rgb(230, 210, 50)",
		"rgb(150, 60, 30)",
		"rgb(70, 10, 20)",
	];
	const radarTicks = [0, 2.5, 5, 7.5, 10];

	const comparisonColor = "rgb(0 145 213)";
	const shiftColor = "rgb(234 105 71)"; // radarAreaSplitColors[Math.max(0, radarTicks.findLastIndex(v => (v < burdenValue)))];

	const burdenRadarOptions = $derived<echarts.EChartsOption>({
		color: [comparisonColor, shiftColor],
		title: {show: false},
		legend: {
			show: true,
			orient: "vertical",
			align: "left",
			left: 0,
			textStyle: {
				color: "white",
			}
		},
		radar: [
			{
				indicator: indicators,
				center: ["50%", "50%"],
				radius: 110,
				shape: "circle",
				axisName: {
					color: "rgb(220, 220, 220)",
					fontSize: 14,
					fontWeight: "normal",
					borderRadius: 3,
					padding: [4, 4],
				},
				axisTick: {
					show: false,
					customValues: radarTicks,
				},
				axisLine: {
					lineStyle: {
						color: "rgb(124 144 154)",
					}
				},
				splitLine: {
					lineStyle: {
						color: "rgb(124 144 154)",
					}
				},
				splitArea: {
					areaStyle: {
						color: ["rgb(35 40 46)"],
						opacity: 0.6,
					}
				},
			},
		],
		series: [
			{
				type: "radar",
				data: [
					{
						value: comparisonStats,
						name: comparisonSetName,
						symbol: "rect",
						symbolSize: 6,
						lineStyle: {
							type: "dashed",
							width: 2,
						},
						label: {
							show: false,
						},
					},
					{
						value: shiftStats,
						name: "This Shift",
						areaStyle: {
							
						},
					},
				],
			},
		],
	});
</script>

<EChart init={echarts.init} options={burdenRadarOptions} />
