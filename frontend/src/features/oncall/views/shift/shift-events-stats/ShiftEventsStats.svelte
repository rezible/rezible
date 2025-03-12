<script lang="ts">
	import { Header, Icon } from "svelte-ux";
	import { mdiAlarmLight, mdiFire, mdiSleepOff } from "@mdi/js";
	import { ZonedDateTime } from "@internationalized/date";
	import ShiftEventsHeatmap from "./ShiftEventsHeatmap.svelte";
	import { cls } from "@layerstack/tailwind";
	import { flatShiftEvents, makeFakeShiftEvents, shiftEventMatchesFilter, type ShiftEvent, type ShiftEventFilterKind } from "$features/oncall/lib/shift";

	type Props = {
		shiftEvents: ShiftEvent[];
		shiftStart: ZonedDateTime;
		shiftEnd: ZonedDateTime;
	};
	const { shiftEvents, shiftStart, shiftEnd }: Props = $props();

	let filterKind = $state<ShiftEventFilterKind>();
	const onEventKindClicked = (kind: ShiftEventFilterKind) => {
		if (filterKind === kind) {
			filterKind = undefined;
			return;
		}
		filterKind = kind;
	};

	const alerts = $derived(shiftEvents.filter(e => e.eventType === "alert"));
	const alertRating = $derived("Normal");

	const nightAlerts = $derived(shiftEvents.filter(e => shiftEventMatchesFilter(e, "nightAlerts")));
	const nightAlertsRating = $derived("Above Average");

	const incidents = $derived(shiftEvents.filter(e => e.eventType === "incident"));
	const incidentsRating = $derived("Above Average");

	const shiftFilteredEvents = $derived(flatShiftEvents(shiftStart, shiftEnd, shiftEvents, filterKind));
</script>

<div class="flex flex-col gap-2 flex-1 min-h-0 max-h-full overflow-y-auto border rounded-lg p-2">
	<Header title="Events" subheading="" classes={{ title: "text-xl" }} />

	<div class="flex flex-row gap-2">
		{#snippet eventTypeBox(kind: ShiftEventFilterKind, label: string, rating: string, icon: string)}
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

	<ShiftEventsHeatmap data={shiftFilteredEvents} />
</div>
