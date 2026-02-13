<script lang="ts">
	import { mdiLineScan, mdiPhoneAlert, mdiMoonWaxingCrescent, mdiClipboardText, mdiCalendarRange } from "@mdi/js";
	import MetricCard from "$src/components/viz/MetricCard.svelte";
	import { useAlertViewState } from "$features/alert/lib/viewState.svelte";
	import { DateRangeField } from "svelte-ux";
	import { makeCalendarDateString, makeDateRangeWindow } from "$lib/date-utils";
	import { getAlertMetricsOptions, type GetAlertMetricsData } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";

	const view = useAlertViewState();
	
	const defaultDateRange = makeDateRangeWindow({ days: 7 });
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
	const actionability = $derived((!!metrics && metrics.feedbacks > 0) ? (metrics.actionable / metrics.feedbacks) : 0);
	const documentation = $derived((!!metrics && metrics.docsAvailable > 0) ? (metrics.docsAvailable / metrics.feedbacks) : 0);
</script>

<div class="flex flex-col gap-2">
	<DateRangeField 
		classes={{field: {root: "w-fit"}}} 
		value={dateRange} 
		on:change={e => (dateRange = e.detail)} 
		label="Date Range"
		icon={mdiCalendarRange}
	/>

	{#if metrics}
		<div class="flex flex-col">
			<h1>Events</h1>
			<div class="flex gap-2 mb-2">
				<MetricCard
					title="Trigger Events"
					icon={mdiLineScan}
					metric={metrics.triggers}
				/>
				<MetricCard
					title="Interrupts"
					icon={mdiPhoneAlert}
					metric={metrics.interrupts}
				/>
				<MetricCard
					title="Night Interrupts"
					icon={mdiMoonWaxingCrescent}
					metric={metrics.nightInterrupts}
				/>
			</div>
			
			<h1>Feedback</h1>
			<div class="flex gap-2">
				<MetricCard
					title="Feedback Given"
					icon={mdiClipboardText}
					metric={metrics.feedbacks}
				/>
				<MetricCard
					title="Actionable"
					icon={mdiClipboardText}
					metric={actionability}
					format="percentage"
				/>
				<MetricCard 
					title="Accurate (Yes/No/Unknown)" 
					icon={mdiClipboardText} 
					metric={accuracy}
				/>
				<MetricCard
					title="Documentation Available"
					icon={mdiClipboardText}
					metric={documentation}
				/>
			</div>
		</div>
	{/if}
</div>