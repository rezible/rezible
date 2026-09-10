<script lang="ts">
	import RiScanLine from "remixicon-svelte/icons/scan-line";
	import RiPhoneFindLine from "remixicon-svelte/icons/phone-find-line";
	import RiMoonLine from "remixicon-svelte/icons/moon-line";
	import RiClipboardLine from "remixicon-svelte/icons/clipboard-line";
	import RiCalendarLine from "remixicon-svelte/icons/calendar-line";
	import { CalendarDate, getLocalTimeZone, now, type DateTimeDuration } from "@internationalized/date";
	import MetricCard from "$src/components/viz/MetricCard.svelte";
	import { useAlertViewController } from "$features/signals/views/signal";
	import { getAlertMetricsOptions, type GetAlertMetricsData } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";

	const view = useAlertViewController();

	const makeCalendarDateString = (d: Date) =>
		new CalendarDate(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate()).toString();

	const defaultDateRange = {
		from: now(getLocalTimeZone()).subtract({ days: 7 }).toDate(),
		to: now(getLocalTimeZone()).toDate(),
		periodType: "day",
	};
	let dateRange = $state({ from: defaultDateRange.from, to: defaultDateRange.to });

	const dateFrom = $derived(!!dateRange?.from ? dateRange.from : defaultDateRange.from);
	const dateTo = $derived(!!dateRange?.to ? dateRange.to : defaultDateRange.to);
	const queryData = $derived<GetAlertMetricsData["query"]>({
		from: makeCalendarDateString(dateFrom),
		to: makeCalendarDateString(dateTo),
	});
	const query = createQuery(() => getAlertMetricsOptions({ path: { id: view.alertId }, query: queryData }));
	const metrics = $derived(query.data?.data);

	const notAccurateFbs = $derived(
		!!metrics ? metrics.feedbacks - metrics.accurate - metrics.accurateUnknown : 0
	);
	const accuracy = $derived(
		!!metrics ? `${metrics.accurate}/${notAccurateFbs}/${metrics.accurateUnknown}` : ""
	);
	const actionability = $derived(
		!!metrics && metrics.feedbacks > 0 ? metrics.actionable / metrics.feedbacks : 0
	);
	const documentation = $derived(
		!!metrics && metrics.docsAvailable > 0 ? metrics.docsAvailable / metrics.feedbacks : 0
	);
</script>

<div class="flex flex-col gap-2">
	<!-- <DateRangeField 
		classes={{field: {root: "w-fit"}}} 
		value={dateRange} 
		onchange={e => (dateRange = e.detail)} 
		label="Date Range"
		icon={RiCalendarLine}
	/> -->

	{#if metrics}
		<div class="flex flex-col">
			<h1>Events</h1>
			<div class="flex gap-2 mb-2">
				<MetricCard title="Trigger Events" icon={RiScanLine} metric={metrics.triggers} />
				<MetricCard title="Interrupts" icon={RiPhoneFindLine} metric={metrics.interrupts} />
				<MetricCard title="Night Interrupts" icon={RiMoonLine} metric={metrics.nightInterrupts} />
			</div>

			<h1>Feedback</h1>
			<div class="flex gap-2">
				<MetricCard title="Feedback Given" icon={RiClipboardLine} metric={metrics.feedbacks} />
				<MetricCard
					title="Actionable"
					icon={RiClipboardLine}
					metric={actionability}
					format="percentage"
				/>
				<MetricCard title="Accurate (Yes/No/Unknown)" icon={RiClipboardLine} metric={accuracy} />
				<MetricCard title="Documentation Available" icon={RiClipboardLine} metric={documentation} />
			</div>
		</div>
	{/if}
</div>
