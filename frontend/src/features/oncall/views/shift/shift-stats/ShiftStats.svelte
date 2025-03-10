<script lang="ts">
	import { Button, Checkbox, Header, Icon, ListItem } from "svelte-ux";
	import { mdiAlarmLight, mdiChevronDown, mdiFilter, mdiFire, mdiSleepOff } from "@mdi/js";
	import type { OncallShift } from "$src/lib/api";
	import { parseAbsoluteToLocal, ZonedDateTime } from "@internationalized/date";
	import ShiftAlertsGraph from "./ShiftAlertsGraph.svelte";
	import { differenceInCalendarDays } from "date-fns";
	import { cls } from "@layerstack/tailwind";

	type Props = {
		shift: OncallShift;
	};
	const { shift }: Props = $props();

	// TODO: use shift timezone
	const shiftStart = $derived(parseAbsoluteToLocal(shift.attributes.startAt));
	const shiftEnd = $derived(parseAbsoluteToLocal(shift.attributes.endAt));

	type ShiftEvent = {
		timestamp: ZonedDateTime;
		eventType: "alert" | "incident";
	};
	const fakeEvent = (day: ZonedDateTime): ShiftEvent => {
		const isAlert = Math.random() > 0.25;
		const timestamp = day
			.copy()
			.set({ hour: Math.floor(Math.random() * 24), minute: Math.floor(Math.random() * 60) });
		return { timestamp, eventType: isAlert ? "alert" : "incident" };
	};
	const dayEvents = (day: number): ShiftEvent[] => {
		const date = shiftStart.add({ days: day });
		const numEvents = Math.floor(Math.random() * 20);
		return Array.from({ length: numEvents }, (_, i) => fakeEvent(date));
	};

	const makeFakeShiftEvents = (start: ZonedDateTime, end: ZonedDateTime) => {
		const shiftDays = differenceInCalendarDays(end.toDate(), start.toDate());
		let events: ShiftEvent[] = [];
		for (let i = 0; i < shiftDays; i++) {
			events = events.concat(dayEvents(i));
		}
		return events;
	};
	const shiftEvents = $derived(makeFakeShiftEvents(shiftStart, shiftEnd));

	type FilterKind = "alerts" | "nightAlerts" | "incidents";
	let filterKind = $state<FilterKind>();

	const matchesFilter = (event: ShiftEvent, kind: FilterKind) => {
		if ((kind === "alerts" || kind === "nightAlerts") && event.eventType !== "alert") return false;
		if (kind === "nightAlerts" && (event.timestamp.hour < 18 && event.timestamp.hour > 6)) return false;
		if (kind === "incidents" && event.eventType !== "incident") return false;
		return true;
	}

	const eventDayKey = (day: number, hour: number) => `${day}-${hour}`;
	const flatEvents = (start: ZonedDateTime, end: ZonedDateTime, events: ShiftEvent[], kind?: FilterKind) => {
		const day1 = start.toDate();

		const numEvents = new Map<string, number>();
		events.forEach((event) => {
			if (!!kind && !matchesFilter(event, kind)) return;
			const day = differenceInCalendarDays(event.timestamp.toDate(), day1);
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

	const onEventKindClicked = (kind: FilterKind) => {
		if (filterKind === kind) {
			filterKind = undefined;
			return;
		}
		filterKind = kind;
	};

	const alerts = $derived(shiftEvents.filter(e => e.eventType === "alert"));
	const alertRating = $derived("Normal");

	const nightAlerts = $derived(shiftEvents.filter(e => matchesFilter(e, "nightAlerts")));
	const nightAlertsRating = $derived("Above Average");

	const incidents = $derived(shiftEvents.filter(e => e.eventType === "incident"));
	const incidentsRating = $derived("Above Average");

	const shiftFilteredEvents = $derived(flatEvents(shiftStart, shiftEnd, shiftEvents, filterKind));
</script>

<div class="flex flex-col gap-2 flex-1 min-h-0 max-h-full overflow-y-auto border rounded-lg p-2">
	<Header title="Events" subheading="" classes={{ title: "text-xl" }} />

	<div class="flex flex-row gap-2">
		{#snippet eventTypeBox(kind: FilterKind, label: string, rating: string, icon: string)}
			{@const isFiltered = filterKind === kind}
			<button
				class={cls(
					"flex gap-4 items-center border-surface-content/10 py-2 relative border-2 px-4 rounded-lg",
					(!!filterKind && isFiltered) && "bg-accent-700/25",
					(!filterKind && !isFiltered) && "bg-surface-100")}
				onclick={() => onEventKindClicked(kind)}
			>
				<div class="flex flex-col">
					<Icon data={icon} />
				</div>
				<div class="flex-grow">
					<span class="text-md text-neutral-content block">{label}</span>
					<span class="text-sm">{rating}</span>
				</div>
			</button>
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

	<ShiftAlertsGraph {shift} data={shiftFilteredEvents} />
</div>
