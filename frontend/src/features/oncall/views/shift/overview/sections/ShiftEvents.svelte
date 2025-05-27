<script lang="ts">
	import { differenceInCalendarDays, getDay } from "date-fns";
	import { Collapse } from "svelte-ux";
	import { mdiBellAlert, mdiBellSleep, mdiFire } from "@mdi/js";

	import type { OncallShiftMetrics } from "$lib/api";
	import { settings } from "$lib/settings.svelte";
	
	import { shiftViewStateCtx } from "../../context.svelte";
	import type { ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import MetricCard from "$components/viz/MetricCard.svelte";
	import ShiftEventsHeatmap from "./ShiftEventsHeatmap.svelte";

	type Props = {
		metrics: OncallShiftMetrics;
		comparison: OncallShiftMetrics;
	};

	let { metrics, comparison }: Props = $props();

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
</script>

<div class="grid grid-cols-3 gap-2">
	<MetricCard
		title="Incidents"
		icon={mdiFire}
		metric={metrics.incidents.total}
		comparison={{value: comparison.incidents.total}}
	/>
	<MetricCard
		title="Alerts"
		icon={mdiBellAlert}
		metric={metrics.alerts.total}
		comparison={{value: comparison.alerts.total}}
	/>
	<MetricCard
		title="Night Alerts"
		icon={mdiBellSleep}
		metric={metrics.alerts.countNight}
		comparison={{value: comparison.alerts.countNight}}
	/>
</div>

<Collapse classes={{ root: "border rounded border-surface-content/10", icon: "mr-2" }}>
	<div slot="trigger" class="flex-1 px-3 py-3">Event Heatmap</div>
	<div class="border-surface-content/10">
		<ShiftEventsHeatmap
			data={hourlyEventCount}
			dayLabels={heatmapDayLabels}
			onDataClicked={onHeatmapHourClicked}
		/>
	</div>
</Collapse>
