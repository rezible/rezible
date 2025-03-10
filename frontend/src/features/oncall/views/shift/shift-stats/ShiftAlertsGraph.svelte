<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import { DayHours } from "$lib/utils.svelte";
	
	import { init, use } from "echarts/core";
	import { HeatmapChart } from "echarts/charts";
	import { GridComponent, TitleComponent, VisualMapComponent, TooltipComponent } from "echarts/components";
	import { CanvasRenderer } from "echarts/renderers";
	import EChart, { type ChartProps } from "$components/echart/EChart.svelte";
	import type { ECMouseEvent } from "$components/echart/events";

	use([HeatmapChart, GridComponent, CanvasRenderer, TitleComponent, VisualMapComponent, TooltipComponent]);

	type Props = {
		shift: OncallShift;
		data: number[][]; // [day, hour, event]
	};
	const { shift, data }: Props = $props();

	const numDays = $derived(data.length / 24);
	const WeekDays = [
		'Saturday', 'Friday', 'Thursday',
		'Wednesday', 'Tuesday', 'Monday', 'Sunday',
	] as const;
	const dayLabels = $derived(Array.from({length: numDays}, (_, i) => WeekDays[i % 7]));

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

	const onClicked = (e: ECMouseEvent) => {
		console.log(e);
	}
</script>

<div class="w-full h-96 p-3 pl-10 bg-surface-100 rounded-lg block overflow-hidden">
	<EChart {init} {options} onclick={onClicked} />
</div>
