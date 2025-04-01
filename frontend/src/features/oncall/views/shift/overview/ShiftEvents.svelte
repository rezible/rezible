<script lang="ts">
	import { differenceInCalendarDays, getDay } from "date-fns";
	import { Collapse } from "svelte-ux";
	import { mdiBellAlert, mdiBellSleep, mdiFire, mdiHeadQuestion } from "@mdi/js";
	import { settings } from "$lib/settings.svelte";
	import type { OncallShiftMetrics } from "$lib/api";
	import { shiftState } from "$features/oncall/views/shift/shift.svelte";
	import { shiftEventMatchesFilter, type ShiftEvent, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import MetricCard from "$components/viz/MetricCard.svelte";
	import ShiftEventsHeatmap from "./ShiftEventsHeatmap.svelte";

	type Props = {
		metrics: OncallShiftMetrics;
		comparison: OncallShiftMetrics;
		shiftEvents: ShiftEvent[];
		eventsFilter: ShiftEventFilterKind | undefined;
	};

	let { metrics, comparison, shiftEvents, eventsFilter = $bindable() }: Props = $props();

	const onEventKindClicked = (kind: ShiftEventFilterKind) => {
		if (eventsFilter === kind) {
			eventsFilter = undefined;
			return;
		}
		eventsFilter = kind;
	};

	const startDate = $derived(shiftState.shiftStart.toDate());
	const endDate = $derived(shiftState.shiftEnd.toDate());

	const eventDayKey = (day: number, hour: number) => `${day}-${hour}`;
	const hourlyEventCount = $derived.by(() => {
		const numDays = differenceInCalendarDays(endDate, startDate);

		const numEvents = new Map<string, number>();
		shiftEvents.forEach((event) => {
			if (!!eventsFilter && !shiftEventMatchesFilter(event, eventsFilter)) return;
			const eventDate = event.timestamp.toDate();
			const day = differenceInCalendarDays(eventDate, startDate);
			const key = eventDayKey(day, event.timestamp.hour);
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
	const numDays = $derived(Math.floor(hourlyEventCount.length / 24));
	const heatmapDayLabels = $derived.by(() => {
		return Array.from({ length: numDays }, (_, day) => {
			const date = shiftState.shiftStart.add({ days: day });
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
</script>

<div class="grid grid-cols-3 gap-2">
	<MetricCard
		title="Incidents"
		icon={mdiFire}
		metric={metrics.totalIncidents}
		comparison={{value: comparison.totalIncidents}}
	/>
	<MetricCard
		title="Alerts"
		icon={mdiBellAlert}
		metric={metrics.totalAlerts}
		comparison={{value: comparison.totalAlerts}}
	/>
	<MetricCard
		title="Night Alerts"
		icon={mdiBellSleep}
		metric={metrics.nightAlerts}
		comparison={{value: comparison.nightAlerts}}
	/>
</div>

<Collapse classes={{ root: "border rounded bg-surface-100/40 border-surface-content/10", icon: "mr-2" }}>
	<div slot="trigger" class="flex-1 px-3 py-3">Events Heatmap</div>
	<div class="border-t border-surface-content/10">
		<ShiftEventsHeatmap
			data={hourlyEventCount}
			dayLabels={heatmapDayLabels}
			onDataClicked={onHeatmapHourClicked}
		/>
	</div>
</Collapse>

<!--div class="flex gap-4">
	<span class="text-lg font-bold">{metrics.totalAlerts}</span>
	<div class="w-[124px] h-6">
		<BarChart
			data={alertHourData}
			x="date"
			y="value"
			axis={false}
			grid={true}
			bandPadding={0.1}
			props={{ bars: { radius: 0, strokeWidth: 0 } }}
		>
			<svelte:fragment slot="tooltip" let:width>
				<Tooltip.Root
					class="text-xs"
					contained={false}
					variant="none"
					y={-10}
					x={width + 8}
					let:data
				>
					<div class="whitespace-nowrap">
						{format(data.date, "eee, MMM do")}
					</div>
					<div class="font-semibold">
						{data.value}
					</div>
				</Tooltip.Root>
			</svelte:fragment>
		</BarChart>
	</div>
</div-->
