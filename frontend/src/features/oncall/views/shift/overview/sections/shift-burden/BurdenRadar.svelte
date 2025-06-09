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
		max: 10,
	}));

	const comparisonStats = burdenStats.map(v => (v.comparison?.value || 0))
	const shiftStats = burdenStats.map(v => (v.value || 0))

	const burdenRadarOptions = $derived<echarts.EChartsOption>({
		color: ["rgba(90, 100, 190, 1)", "rgba(170, 110, 20, 1)"],
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
					borderRadius: 3,
					padding: [4, 4],
				},
				splitArea: {
					interval: 20,
					areaStyle: {
						color: ["rgba(150, 150, 190, .3)", "rgba(30, 250, 30, .2)", 'rgba(230, 210, 50, .2)', 'rgba(150, 90, 20, .3)', 'rgba(180, 30, 40, .2)'],
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
							
						},
					},
				],
			},
		],
	});
</script>

<EChart init={echarts.init} options={burdenRadarOptions} />
