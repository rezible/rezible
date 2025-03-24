<script lang="ts">
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import { simulateApiDelay, type OncallShift } from "$lib/api";
	import {
		makeFakeComparisonMetrics,
		makeFakeShiftMetrics,
		type ComparisonMetrics,
		type ShiftMetrics,
	} from "$features/oncall/lib/shift-metrics";
	import {
		buildShiftTimeDetails,
		formatShiftDates,
		shiftEventMatchesFilter,
		type ShiftEvent,
		type ShiftEventFilterKind,
	} from "$features/oncall/lib/utils";

	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";

	import ShiftEvents from "./shift-events/ShiftEvents.svelte";
	import IncidentMetrics from "./shift-details/IncidentMetrics.svelte";
	import WorkloadBreakdown from "./shift-details/WorkloadBreakdown.svelte";
	import ShiftEventsList from "./shift-events/ShiftEventsList.svelte";
	import { getLocalTimeZone, parseAbsolute } from "@internationalized/date";

	type Props = {
		shift: OncallShift;
		events: ShiftEvent[];
	};
	let { shift, events }: Props = $props();

	const shiftTimeDetails = buildShiftTimeDetails(shift);

	// TODO: default to shift timezone & allow choosing timezone
	let eventTimezone = $state(getLocalTimeZone());
	const shiftStart = $derived(parseAbsolute(shift.attributes.startAt, eventTimezone));
	const shiftEnd = $derived(parseAbsolute(shift.attributes.endAt, eventTimezone));

	const shiftMetricsQuery = createQuery(() =>
		queryOptions({
			queryKey: ["shiftMetrics", shift.id],
			queryFn: async (): Promise<{ data: ShiftMetrics }> => {
				await simulateApiDelay(500);
				return { data: makeFakeShiftMetrics() };
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);
	const metrics = $derived(shiftMetricsQuery.data?.data);

	const shiftComparisonQuery = createQuery(() =>
		queryOptions({
			queryKey: ["shiftComparison", shift.id],
			queryFn: async (): Promise<{ data: ComparisonMetrics }> => {
				await simulateApiDelay(500);
				return { data: makeFakeComparisonMetrics() };
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);
	const comparison = $derived(shiftComparisonQuery.data?.data);

	let eventsFilter = $state<ShiftEventFilterKind>();
	const shiftEvents = $derived(
		!eventsFilter
			? events
			: events.filter((e) => !eventsFilter || shiftEventMatchesFilter(e, eventsFilter))
	);
</script>

<div class="w-full h-full grid grid-cols-3 gap-2">
	{#if !metrics || !comparison}
		<div class="grid col-span-3 w-full h-full place-items-center">
			<LoadingIndicator />
		</div>
	{:else}
		<div class="col-span-2 h-full w-full overflow-y-auto space-y-2">
			<ShiftEvents {shift} {shiftEvents} {metrics} {comparison} bind:eventsFilter />

			<WorkloadBreakdown {shift} {shiftEvents} {metrics} />

			<IncidentMetrics {metrics} {comparison} />
		</div>

		<div class="h-full flex flex-col overflow-y-auto">
			<ShiftEventsList {shiftStart} {shiftEvents} />
		</div>
	{/if}
</div>
