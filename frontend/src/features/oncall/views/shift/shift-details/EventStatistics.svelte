<script lang="ts">
	import { Card, Header } from "svelte-ux";
	import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { formatDelta } from "$lib/format.svelte";
	import { BarChart, Tooltip } from "layerchart";
	import type { ShiftEvent } from "$features/oncall/lib/utils";
	import { differenceInHours, format, roundToNearestHours } from "date-fns";
	import type { OncallShift } from "$src/lib/api";
	import { getLocalTimeZone, parseAbsolute, ZonedDateTime } from "@internationalized/date";
	import { settings } from "$src/lib/settings.svelte";

	type Props = {
		shift: OncallShift;
		metrics: ShiftMetrics;
		shiftEvents: ShiftEvent[];
		comparison: ComparisonMetrics;
	};

	let { shift, metrics, shiftEvents, comparison }: Props = $props();

	const comparisonClass = (value: number) => {
		if (value > 0) return "text-red-500";
		if (value < 0) return "text-green-500";
		return "text-gray-500";
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
</script>

<div class="flex flex-col gap-2 border p-2">
	<Header title="Event Statistics" subheading="events" />

	<div class="flex divide-x gap-4 p-2">
		<div class="flex flex-col px-4">
			<span>Alerts</span>
			<div class="flex gap-4">
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
			</div>
			<div class={comparisonClass(comparison.alertsComparison)}>
				{formatDelta(comparison.alertsComparison)} from average
			</div>
		</div>

		<div class="flex flex-col px-4">
			<span>Night Alerts</span>
			<span class="text-lg font-bold">{metrics.nightAlerts}</span>
			<div class={comparisonClass(comparison.nightAlertsComparison)}>
				{formatDelta(comparison.nightAlertsComparison)} from average
			</div>
			<div class="text-sm text-gray-500 mt-1">Potential sleep disruptions</div>
		</div>

		<div class="flex flex-col px-4">
			<span>Incidents</span>
			<span class="text-lg font-bold">{metrics.totalIncidents}</span>
			<div class={comparisonClass(comparison.incidentsComparison)}>
				{formatDelta(comparison.incidentsComparison)} from average
			</div>
		</div>
	</div>
</div>
