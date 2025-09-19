<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getOncallShiftMetricsOptions } from "$lib/api";

	import ShiftSummaryCards from "./sections/ShiftSummaryCards.svelte";
	import ShiftIncidents from "./sections/ShiftIncidents.svelte";
	import ShiftBurden from "./sections/shift-burden/ShiftBurden.svelte";
	import ShiftAlerts from "./sections/ShiftAlerts.svelte";

	import ShiftEventsSidebar from "./events-sidebar/ShiftEventsSidebar.svelte";
	import EventAnnotationDialog from "$src/components/events/annotation-dialog/EventAnnotationDialog.svelte";
	import { useOncallShiftViewState } from "$features/oncall-shift";

	const view = useOncallShiftViewState();
	const shiftId = $derived(view.shiftId);

	const comparisonQuery = createQuery(() => getOncallShiftMetricsOptions());
	const comparison = $derived(comparisonQuery.data?.data);
	
	const metricsQuery = createQuery(() => getOncallShiftMetricsOptions({query: {shiftId}}));
	const metrics = $derived(metricsQuery.data?.data);
</script>

<div class="w-full h-full grid grid-cols-3 gap-2">
	<div class="col-span-2 h-full w-full overflow-y-auto pr-1 space-y-2">
		<ShiftSummaryCards {metrics} {comparison} />
		<ShiftBurden {metrics} {comparison} />
		<ShiftAlerts {metrics} />
		<ShiftIncidents {metrics} {comparison} />
	</div>

	<div class="h-full flex flex-col overflow-y-auto">
		<ShiftEventsSidebar />
	</div>
</div>

<EventAnnotationDialog />