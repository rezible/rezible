<script lang="ts">
	import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { formatDelta } from "$lib/format.svelte";
	import {
		shiftEventMatchesFilter,
		type ShiftEvent,
		type ShiftEventFilterKind,
	} from "$features/oncall/lib/utils";
	import { differenceInHours, roundToNearestHours, differenceInCalendarDays, getDay } from "date-fns";
	import type { OncallShift } from "$lib/api";
	import { getLocalTimeZone, parseAbsolute, ZonedDateTime } from "@internationalized/date";
	import { settings } from "$lib/settings.svelte";
	import { Collapse, Icon } from "svelte-ux";
	import ShiftEventsHeatmap from "./ShiftEventsHeatmap.svelte";
	import { mdiAlarm, mdiBellAlert, mdiBellSleep, mdiFire, mdiSleepOff } from "@mdi/js";

	type Props = {
		shift: OncallShift;
		metrics: ShiftMetrics;
		comparison: ComparisonMetrics;
		shiftEvents: ShiftEvent[];
		eventsFilter: ShiftEventFilterKind | undefined;
	};

	let { shift, metrics, comparison, shiftEvents, eventsFilter = $bindable() }: Props = $props();

	const comparisonClass = (value: number) => {
		if (value > 0) return "text-red-500";
		if (value < 0) return "text-green-500";
		return "text-gray-500";
	};

	const onEventKindClicked = (kind: ShiftEventFilterKind) => {
		if (eventsFilter === kind) {
			eventsFilter = undefined;
			return;
		}
		eventsFilter = kind;
	};

	const hourBucketSize = 6;

	const bucketDate = (d: ZonedDateTime) => {
		const rounded = d.set({ minute: 0, second: 0, millisecond: 0 });
		const diff = rounded.hour % hourBucketSize;
		return rounded.subtract({ hours: diff });
	};

	// TODO: use shift state value
	const shiftStart = $derived(parseAbsolute(shift.attributes.startAt, getLocalTimeZone()));
	const shiftEnd = $derived(parseAbsolute(shift.attributes.endAt, getLocalTimeZone()));

	// const alerts = $derived(events.filter((e) => e.eventType === "alert"));
	// const nightAlerts = $derived(events.filter((e) => shiftEventMatchesFilter(e, "nightAlerts")));
	// const incidents = $derived(events.filter((e) => e.eventType === "incident"));

	const alerts = $derived(shiftEvents.filter((e) => e.eventType === "alert"));
	const alertHourData = $derived.by(() => {
		const counts = new Map<string, number>();
		alerts.forEach((ev) => {
			const key = bucketDate(ev.timestamp).toAbsoluteString();
			counts.set(key, (counts.get(key) || 0) + 1);
		});
		const startHour = shiftStart.set({ minute: 0, second: 0, millisecond: 0 });
		const buckets =
			differenceInHours(roundToNearestHours(shiftEnd.toDate()), startHour.toDate()) / hourBucketSize;

		return Array.from({ length: buckets }, (_, bucket) => {
			const d = bucketDate(startHour.add({ hours: bucket * hourBucketSize }));
			const key = d.toAbsoluteString();
			return { date: d.toDate(), value: counts.get(key) || 0 };
		});
	});

	const eventDayKey = (day: number, hour: number) => `${day}-${hour}`;
	const formatShiftEventCountForHeatmap = (
		start: ZonedDateTime,
		end: ZonedDateTime,
		events: ShiftEvent[],
		kind?: ShiftEventFilterKind
	) => {
		const startDate = start.toDate();

		const numEvents = new Map<string, number>();
		events.forEach((event) => {
			if (!!kind && !shiftEventMatchesFilter(event, kind)) return;
			const eventDate = event.timestamp.toDate();
			const day = differenceInCalendarDays(eventDate, startDate);
			const key = eventDayKey(day, event.timestamp.hour);
			numEvents.set(key, (numEvents.get(key) || 0) + 1);
		});

		const numDays = differenceInCalendarDays(end.toDate(), start.toDate());

		return Array.from({ length: numDays }).flatMap((_, day) => {
			return Array.from({ length: 24 }).map((_, hour) => [
				day,
				hour,
				numEvents.get(eventDayKey(day, hour)) || 0,
			]);
		});
	};

	const hourlyEventCount = $derived(
		formatShiftEventCountForHeatmap(shiftStart, shiftEnd, shiftEvents, eventsFilter)
	);
	const numDays = $derived(Math.floor(hourlyEventCount.length / 24));
	const heatmapDayLabels = $derived.by(() => {
		const fmt = settings.format;
		return Array.from({ length: numDays }, (_, day) => {
			const date = shiftStart.add({ days: day });
			const dayOfWeek = getDay(date.toAbsoluteString());
			const dayName = fmt.getDayOfWeekName(dayOfWeek);
			return `${dayName} ${String(date.day).padStart(2, "0")}`;
		});
	});

	const onHeatmapHourClicked = (idx: number) => {
		if (idx < 0 || idx > hourlyEventCount.length) return;
		const [day, hour] = hourlyEventCount[idx];
		console.log(day, hour);
	};
</script>

{#snippet statBox(title: string, icon: string, metric: number, comparison?: number, hint?: string)}
	<div class="flex flex-col border rounded py-3 px-4 bg-surface-100/40 border-surface-content/10">
		<div class="w-full flex justify-between items-center">
			<span class="">{title}</span>
			<span class=""><Icon data={icon} /></span>
		</div>
		<span class="text-lg font-bold">{metric}</span>
		{#if comparison}
			<div class="{comparisonClass(comparison)} flex items-center gap-2">
				{formatDelta(comparison)} from average

				{#if hint}
				<div class="text-warning">
					<Icon data={mdiSleepOff} size={16} classes={{root: "border rounded-full border-warning"}} />
					<!--div class="text-sm text-gray-500 mt-1">
						Potential sleep disruptions
					</div-->
				</div>
				{/if}
			</div>
		{/if}
	</div>
{/snippet}

<div class="grid grid-cols-3 gap-2">
	{@render statBox("Alerts", mdiBellAlert, metrics.totalAlerts, comparison.alertsComparison)}
	{@render statBox("Night Alerts", mdiBellSleep, metrics.nightAlerts, comparison.nightAlertsComparison, "Potential sleep disruptions")}
	{@render statBox("Incidents", mdiFire, metrics.totalIncidents, comparison.incidentsComparison)}
</div>

<Collapse classes={{root: "border rounded bg-surface-100/40 border-surface-content/10", icon: "mr-2"}}>
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