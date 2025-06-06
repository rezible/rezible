<script lang="ts">
	import * as echarts from "echarts";
	import EChart from "$components/viz/echart/EChart.svelte";
	import type { InlineStatProps } from "$components/viz/InlineStat.svelte";

	type Props = {
		burdenStats: InlineStatProps[];
		comparisonSetName?: string;
	};
	const { burdenStats, comparisonSetName = "Roster Average" }: Props = $props();

	const indicators = burdenStats.map(v => ({
		text: v.title.replaceAll(" ", "\n"),
		min: 0,
		max: 100,
	}));

	const comparisonStats = burdenStats.map(v => (v.comparison?.value || 0))
	const shiftStats = burdenStats.map(v => (v.value || 0))

	const burdenRadarOptions = $derived<echarts.EChartsOption>({
		color: ["rgba(100, 10, 120, 1)", "#FFE434"],
		title: {show: false},
		legend: {
			show: true,
			textStyle: {
				color: "white",
			}
		},
		radar: [
			{
				indicator: indicators,
				center: ["50%", "60%"],
				radius: 110,
				axisName: {
					color: "#fff",
					backgroundColor: "rgb(10, 10, 10)",
					borderRadius: 3,
					padding: [4, 4],
				},
				splitArea: {
					interval: 20,
					areaStyle: {
						color: ["rgba(130, 130, 160, .9)", "rgba(30, 100, 30, 1)", 'rgba(160, 130, 50, 1)', 'rgba(140, 50, 20, .2)', 'rgba(180, 30, 40, .2)'],
						shadowColor: 'rgba(0, 0, 0, 0.2)',
						shadowBlur: 10
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
							color: new echarts.graphic.RadialGradient(0.1, 0.6, 1, [
								{
									color: "rgba(255, 145, 124, 0.1)",
									offset: 0,
								},
								{
									color: "rgba(255, 145, 124, 0.9)",
									offset: 1,
								},
							]),
						},
					},
				],
			},
		],
	});
</script>

<EChart init={echarts.init} options={burdenRadarOptions} />
