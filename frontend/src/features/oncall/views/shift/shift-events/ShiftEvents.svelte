<script lang="ts">
	import { Button, Header, Icon } from "svelte-ux";
	import { mdiAlarmLight, mdiFire, mdiSleepOff } from "@mdi/js";
	import { ZonedDateTime } from "@internationalized/date";
	import ShiftEventsHeatmap from "./ShiftEventsHeatmap.svelte";
	import { cls } from "@layerstack/tailwind";
	import { shiftEventMatchesFilter, type ShiftEvent, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import { differenceInCalendarDays, getDay } from "date-fns";
	import { settings } from "$src/lib/settings.svelte";
	import ShiftEventsList from "./ShiftEventsList.svelte";

	type Props = {
		events: ShiftEvent[];
		shiftStart: ZonedDateTime;
		shiftEnd: ZonedDateTime;
	};
	const { events, shiftStart, shiftEnd }: Props = $props();

	let eventsFilter = $state<ShiftEventFilterKind>();
	const filteredEvents = $derived(events.filter(e => !eventsFilter || shiftEventMatchesFilter(e, eventsFilter)));

	const onEventKindClicked = (kind: ShiftEventFilterKind) => {
		if (eventsFilter === kind) {
			eventsFilter = undefined;
			return;
		}
		eventsFilter = kind;
	};

	// TODO: Implement this properly
	const getEventsRating = (count: number) => {
		if (count === 0) return "Normal";
		if (count < 3) return "Below Average";
		if (count < 6) return "Normal";
		if (count < 9) return "Above Average";
		return "High";
	};

	const eventDayKey = (day: number, hour: number) => `${day}-${hour}`;
	const formatShiftEventCountForHeatmap = (start: ZonedDateTime, end: ZonedDateTime, events: ShiftEvent[], kind?: ShiftEventFilterKind) => {
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

	const alerts = $derived(events.filter(e => e.eventType === "alert"));
	const alertRating = $derived(getEventsRating(alerts.length));

	const nightAlerts = $derived(events.filter(e => shiftEventMatchesFilter(e, "nightAlerts")));
	const nightAlertsRating = $derived(getEventsRating(nightAlerts.length));

	const incidents = $derived(events.filter(e => e.eventType === "incident"));
	const incidentsRating = $derived(getEventsRating(incidents.length));

	const hourlyEventCount = $derived(formatShiftEventCountForHeatmap(shiftStart, shiftEnd, events, eventsFilter));
	const numDays = $derived(Math.floor(hourlyEventCount.length / 24));
	const heatmapDayLabels = $derived.by(() => {
		const fmt = settings.format;
		return Array.from({length: numDays}, (_, day) => {
			const date = shiftStart.add({days: day});
			const dayOfWeek = getDay(date.toAbsoluteString());
			const dayName = fmt.getDayOfWeekName(dayOfWeek);
			return `${dayName} ${String(date.day).padStart(2, "0")}`;
		});
	});

	const onHeatmapHourClicked = (idx: number) => {
		if (idx < 0 || idx > hourlyEventCount.length) return;
		const [day, hour] = hourlyEventCount[idx];
		console.log(day, hour);
	}
</script>

<div class="flex gap-2 flex-1 h-full overflow-y-auto">
	<div class="flex flex-col gap-2 flex-1 min-h-0 max-h-full overflow-y-auto">
		<div class="">
			<Header title="Events" subheading="Select a filter below to view specific event types" classes={{ title: "text-xl" }}>
				<div slot="actions" class="text-sm text-surface-600 mb-1 flex items-center">
					{#if eventsFilter}
						<Button variant="fill-light"
							on:click={() => eventsFilter = undefined}
						>
							Clear filter
						</Button>
					{:else}
						<span class="">No filter applied</span>
					{/if}
				</div>
			</Header>

			<div class="grid grid-cols-3 gap-2 auto-rows-min mt-2">	
				{#snippet eventTypeBox(kind: ShiftEventFilterKind, label: string, rating: string, icon: string)}
					{@const isFiltered = eventsFilter === kind}
					{@const backgroundCol = rating === "High" ? "bg-warning-400/20" : "bg-surface-100"}
					<div class="grid">
						<button
							class={cls(
								"flex gap-4 items-center py-2 relative rounded-lg border",
								(!!eventsFilter && isFiltered) && "bg-accent-700/25 border-accent-700",
								(!eventsFilter && !isFiltered) && backgroundCol)}
							onclick={() => onEventKindClicked(kind)}
						>
							<div class="flex-grow flex items-center justify-center gap-4">
								<div class="flex flex-col">
									<Icon data={icon} size={28} />
								</div>
								<div class="">
									<span class="text-md text-neutral-content block">{label}</span>
									<span class="text-sm">{rating}</span>
								</div>
							</div>
						</button>
					</div>
				{/snippet}

				{@render eventTypeBox("alerts", `${alerts.length} Alerts`, alertRating, mdiAlarmLight)}
				{@render eventTypeBox(
					"nightAlerts",
					`${nightAlerts.length} Alerts at Night`,
					nightAlertsRating,
					mdiSleepOff
				)}
				{@render eventTypeBox("incidents", `${incidents.length} Incidents`, incidentsRating, mdiFire)}
			</div>
		</div>

		<ShiftEventsHeatmap data={hourlyEventCount} dayLabels={heatmapDayLabels} onDataClicked={onHeatmapHourClicked} />
	</div>

	<div class="w-1/3">
		<ShiftEventsList shiftEvents={filteredEvents} {shiftStart} />
	</div>
</div>