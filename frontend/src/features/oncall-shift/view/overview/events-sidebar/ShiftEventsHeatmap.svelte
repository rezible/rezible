<script lang="ts">
	import { DayHours } from "$lib/utils.svelte";
	
	import { init, use } from "echarts/core";
	import { HeatmapChart } from "echarts/charts";
	import { GridComponent, TitleComponent, VisualMapComponent, TooltipComponent } from "echarts/components";
	import { CanvasRenderer } from "echarts/renderers";
	import EChart, { type ChartProps, type ECMouseEvent } from "$components/viz/echart/EChart.svelte";
	import type { ShiftEventFilterKind } from "$features/oncall-shift/lib/utils";
	import { differenceInCalendarDays, getDay } from "date-fns";
	import { settings } from "$lib/settings.svelte";
	import type { YAXisOption, XAXisOption } from "echarts/types/dist/shared";
	import { useOncallShiftViewState } from "$features/oncall-shift";

	use([HeatmapChart, GridComponent, CanvasRenderer, TitleComponent, VisualMapComponent, TooltipComponent]);

	type Props = {
		onHourClicked: (day: number, hour: number) => void;
	};
	const { onHourClicked }: Props = $props();

	const view = useOncallShiftViewState();

	const onEventKindClicked = (kind: ShiftEventFilterKind) => {
		if (view.eventsFilter === kind) {
			view.eventsFilter = undefined;
			return;
		}
		view.eventsFilter = kind;
	};

	const startDate = $derived(view.shiftStart?.toDate());
	const endDate = $derived(view.shiftEnd?.toDate());

	const numDays = $derived((!!startDate && !!endDate) ? differenceInCalendarDays(endDate, startDate) : 0);

	const eventDayKey = (day: number, hour: number) => `${day}-${hour}`;
	const hourlyEventCount = $derived.by(() => {
		if (!startDate || !endDate) return [];

		const numEvents = new Map<string, number>();
		view.filteredEvents.forEach((event) => {
			const eventDate = new Date(event.attributes.timestamp);
			const day = differenceInCalendarDays(eventDate, startDate);
			const key = eventDayKey(day, eventDate.getHours());
			numEvents.set(key, (numEvents.get(key) || 0) + 1);
		});

		return Array.from({ length: numDays }).flatMap((_, day) => {
			return Array.from({ length: 24 }).map((_, hour) => [
				day,
				hour,
				numEvents.get(eventDayKey(day, hour)) || 0,
			]);
		});
	});
	
	const weekdayLabels = $derived.by(() => {
		return Array.from({ length: numDays }, (_, day) => {
			const start = view.shiftStart;
			if (!start) return "";
			const date = start.add({ days: day });
			const dayOfWeek = getDay(date.toAbsoluteString());
			const dayName = settings.format.getDayOfWeekName(dayOfWeek);
			return `${dayName} ${String(date.day).padStart(2, "0")}`;
		});
	});

	const vertical = true;

	type DayHourCountData = [number, number, number | string];
	const mapDataHorizontalFn = (d: number[]): DayHourCountData => ([d[1], d[0], d[2] || "-"]);
	const mapDataVerticalFn = (d: number[]): DayHourCountData => ([d[0], d[1], d[2] || "-"]);
	const nonZeroData = $derived(hourlyEventCount.map(vertical ? mapDataVerticalFn : mapDataHorizontalFn));

	const weekdaysXAxis = $derived<XAXisOption>({
		position: "top",
		axisLabel: {
			fontSize: 16,
		},
		type: "category",
		inverse: false,
		data: weekdayLabels,
		splitArea: {show: true},
	});

	const hoursYAxis = $derived<YAXisOption>({
		position: "left",
		type: "category",
		data: DayHours,
		inverse: true,
		nameTextStyle: {
			fontSize: 24,
		},
	})

	const options = $derived<ChartProps["options"]>({
		tooltip: {
			position: "top",
			formatter(params) {
				// TODO
				return "day hour - count"
			},
		},
		grid: {
			containLabel: true,
			left: 0,
			top: 0,
			right: 0,
			bottom: 0,
		},
		yAxis: hoursYAxis,
		xAxis: weekdaysXAxis,
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
		const [day, hour] = nonZeroData[e.dataIndex];
		onHourClicked(day, hour)
	};
</script>

<div class="w-full h-full p-3 pl-10 rounded-lg block overflow-hidden">
	<EChart {init} {options} onclick={onClicked} />
</div>