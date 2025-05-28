<script lang="ts">
	import { DayHours } from "$lib/utils.svelte";
	
	import { init, use } from "echarts/core";
	import { HeatmapChart } from "echarts/charts";
	import { GridComponent, TitleComponent, VisualMapComponent, TooltipComponent } from "echarts/components";
	import { CanvasRenderer } from "echarts/renderers";
	import EChart, { type ChartProps, type ECMouseEvent } from "$components/viz/echart/EChart.svelte";
	import { Collapse, Header } from "svelte-ux";
	import { shiftViewStateCtx } from "../../context.svelte";
	import type { ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import { differenceInCalendarDays, getDay } from "date-fns";
	import { settings } from "$lib/settings.svelte";
	import SectionCard from "./SectionCard.svelte";

	use([HeatmapChart, GridComponent, CanvasRenderer, TitleComponent, VisualMapComponent, TooltipComponent]);

	type Props = {
		onDayClicked: (idx: number) => void;
	};
	const { onDayClicked }: Props = $props();

	const viewState = shiftViewStateCtx.get();

	const onEventKindClicked = (kind: ShiftEventFilterKind) => {
		if (viewState.eventsFilter === kind) {
			viewState.eventsFilter = undefined;
			return;
		}
		viewState.eventsFilter = kind;
	};

	const startDate = $derived(viewState.shiftStart?.toDate());
	const endDate = $derived(viewState.shiftEnd?.toDate());

	const numDays = $derived((!!startDate && !!endDate) ? differenceInCalendarDays(endDate, startDate) : 0);

	const eventDayKey = (day: number, hour: number) => `${day}-${hour}`;
	const hourlyEventCount = $derived.by(() => {
		if (!startDate || !endDate) return [];

		const numEvents = new Map<string, number>();
		viewState.filteredEvents.forEach((event) => {
			// if (!!shiftState.eventsFilter && !shiftEventMatchesFilter(event, shiftState.eventsFilter)) return;
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
	
	const heatmapDayLabels = $derived.by(() => {
		return Array.from({ length: numDays }, (_, day) => {
			const start = viewState.shiftStart;
			if (!start) return "";
			const date = start.add({ days: day });
			const dayOfWeek = getDay(date.toAbsoluteString());
			const dayName = settings.format.getDayOfWeekName(dayOfWeek);
			return `${dayName} ${String(date.day).padStart(2, "0")}`;
		});
	});

	const onHeatmapHourClicked = (idx: number) => {
		if (idx < 0 || idx > hourlyEventCount.length) return;
		const [day, hour] = hourlyEventCount[idx];
		console.log(day, hour);
	};

	const nonZeroData = $derived(hourlyEventCount.map(d => ([d[1], d[0], d[2] || "-"])));

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
			data: heatmapDayLabels,
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

	const onClicked = (e: ECMouseEvent) => (onDayClicked(e.dataIndex));
</script>

<SectionCard>
	<Collapse classes={{ root: "", icon: "mr-2" }}>
		<Header title="Event Heatmap" subheading="Alerts by time of day" slot="trigger" class="flex-1" />
		<div class="border-surface-content/10">
			<div class="w-full h-96 p-3 pl-10 rounded-lg block overflow-hidden">
				<EChart {init} {options} onclick={onClicked} />
			</div>
		</div>
	</Collapse>
</SectionCard>