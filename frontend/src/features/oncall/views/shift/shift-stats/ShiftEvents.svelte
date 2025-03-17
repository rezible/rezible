<script lang="ts">
	import { Header, Icon } from "svelte-ux";
	import { mdiAlarmLight, mdiFire, mdiSleepOff } from "@mdi/js";
	import { ZonedDateTime } from "@internationalized/date";
	import ShiftEventsHeatmap from "./ShiftEventsHeatmap.svelte";
	import { cls } from "@layerstack/tailwind";
	import { formatShiftEventCountForHeatmap, shiftEventMatchesFilter, type ShiftEvent, type ShiftEventFilterKind } from "$features/oncall/lib/utils";
	import { getDay } from "date-fns";
	import { settings } from "$src/lib/settings.svelte";
	import { PeriodType } from "@layerstack/utils";

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

	// TODO: Implement this properly
	const getEventsRating = (count: number) => {
		if (count === 0) return "Normal";
		if (count < 3) return "Below Average";
		if (count < 6) return "Normal";
		if (count < 9) return "Above Average";
		return "High";
	};

	const alerts = $derived(shiftEvents.filter(e => e.eventType === "alert"));
	const alertRating = $derived(getEventsRating(alerts.length));

	const nightAlerts = $derived(shiftEvents.filter(e => shiftEventMatchesFilter(e, "nightAlerts")));
	const nightAlertsRating = $derived(getEventsRating(nightAlerts.length));

	const incidents = $derived(shiftEvents.filter(e => e.eventType === "incident"));
	const incidentsRating = $derived(getEventsRating(incidents.length));

	const hourlyEventCount = $derived(formatShiftEventCountForHeatmap(shiftStart, shiftEnd, shiftEvents, filterKind));
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

<div class="flex flex-col gap-2 flex-1 min-h-0 max-h-full overflow-y-auto border rounded-lg p-2">
	<div class="">
		<Header title="Events" subheading="Select a filter below to view specific event types" classes={{ title: "text-xl" }} />

		<div class="grid grid-cols-3 gap-2 auto-rows-min mt-2">
			<div class="col-span-3 text-sm text-surface-600 mb-1 flex items-center">
				<span class="font-medium">Filter events:</span>
				{#if filterKind}
					<button 
						class="ml-auto text-xs bg-surface-200 hover:bg-surface-300 px-2 py-1 rounded-md flex items-center"
						onclick={() => filterKind = undefined}
					>
						Clear filter
					</button>
				{:else}
					<span class="ml-auto text-xs italic">No filter applied</span>
				{/if}
			</div>
			
			{#snippet eventTypeBox(kind: ShiftEventFilterKind, label: string, rating: string, icon: string)}
				{@const isFiltered = filterKind === kind}
				{@const backgroundCol = rating === "High" ? "bg-warning-400/20" : "bg-surface-100"}
				<div class="grid">
					<button
						class={cls(
							"flex gap-4 items-center py-2 relative rounded-lg border",
							(!!filterKind && isFiltered) && "bg-accent-700/25 border-accent-700",
							(!filterKind && !isFiltered) && backgroundCol)}
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
						{#if isFiltered}
							<div class="absolute top-1 right-1 bg-accent-700 text-white text-xs px-1 rounded">
								Filtered
							</div>
						{:else}
							<div class="absolute top-1 right-1 text-xs px-1 rounded opacity-70">
								Click to filter
							</div>
						{/if}
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

	<div class="relative">
		{#if filterKind}
			<div class="absolute -top-6 left-0 text-sm font-medium text-accent-700">
				Showing heatmap for: {filterKind === "alerts" ? "All Alerts" : filterKind === "nightAlerts" ? "Night Alerts" : "Incidents"}
			</div>
		{:else}
			<div class="absolute -top-6 left-0 text-sm text-surface-600">
				Showing heatmap for: All Events
			</div>
		{/if}
		<ShiftEventsHeatmap data={hourlyEventCount} dayLabels={heatmapDayLabels} onDataClicked={onHeatmapHourClicked} />
	</div>
</div>
