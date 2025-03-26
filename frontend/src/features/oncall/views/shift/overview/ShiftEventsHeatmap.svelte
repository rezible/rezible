<script lang="ts">
	import { DayHours } from "$lib/utils.svelte";
	
	import { init, use } from "echarts/core";
	import { HeatmapChart } from "echarts/charts";
	import { GridComponent, TitleComponent, VisualMapComponent, TooltipComponent } from "echarts/components";
	import { CanvasRenderer } from "echarts/renderers";
	import EChart, { type ChartProps, type ECMouseEvent } from "$components/viz/echart/EChart.svelte";

	use([HeatmapChart, GridComponent, CanvasRenderer, TitleComponent, VisualMapComponent, TooltipComponent]);

	type Props = {
		dayLabels: string[];
		data: number[][]; // [day, hour, event]
		onDataClicked: (idx: number) => void;
	};
	const { data, dayLabels, onDataClicked }: Props = $props();

	const nonZeroData = $derived(data.map(d => ([d[1], d[0], d[2] || "-"])));

	const options = $derived<ChartProps["options"]>({
		tooltip: {
			position: "left",
		},
		grid: {
			containLabel: true,
			left: 0,
			top: 0,
			right: 0,
			bottom: 0,
		},
		xAxis: {
			position: "top",
			type: "category",
			data: DayHours,
			nameTextStyle: {
				fontSize: 24,
			},
		},
		yAxis: {
			axisLabel: {
				fontSize: 16,
			},
			type: "category",
			inverse: true,
			data: dayLabels,
			splitArea: {show: true},
		},
		visualMap: {
			min: 0,
			max: 10,
			calculable: false,
			hoverLink: true,
			align: "auto",
			show: false,
			orient: "vertical",
			right: "12px",
			top: "center",
		},
		series: [
			{
				type: "heatmap",
				data: nonZeroData,
				label: {
					show: true,
				},
				emphasis: {
					itemStyle: {
						shadowBlur: 10,
						shadowColor: "rgba(0, 0, 0, 0.5)",
					},
				},
			},
		],
	});

	const onClicked = (e: ECMouseEvent) => (onDataClicked(e.dataIndex));
</script>

<div class="w-full h-96 p-3 pl-10 rounded-lg block overflow-hidden">
	<EChart {init} {options} onclick={onClicked} />
</div>
