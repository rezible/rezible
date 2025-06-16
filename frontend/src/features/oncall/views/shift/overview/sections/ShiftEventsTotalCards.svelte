<script lang="ts">
	import { mdiBellAlert, mdiBellSleep, mdiClockTimeFive, mdiFire } from "@mdi/js";

	import MetricCard from "$components/viz/MetricCard.svelte";
	import type { OncallShiftMetrics } from "$lib/api";

	type Props = {
		metrics?: OncallShiftMetrics;
		comparison?: OncallShiftMetrics;
	};

	let { metrics, comparison }: Props = $props();

	$inspect(metrics, comparison);
</script>

<div class="grid grid-flow-col gap-2">
	<MetricCard
		title="Incidents"
		icon={mdiFire}
		metric={metrics?.incidents.total || 0}
		comparison={{value: comparison?.incidents.total || 0}}
	/>
	<MetricCard
		title="Time in Incidents"
		icon={mdiClockTimeFive}
		metric={metrics?.incidents.responseTimeMinutes || 0}
		format="duration"
		comparison={{value: comparison?.incidents.responseTimeMinutes || 0}}
	/>
	<MetricCard
		title="Alerts"
		icon={mdiBellAlert}
		metric={metrics?.alerts.total || 0}
		comparison={{value: comparison?.alerts.total || 0}}
	/>
	<MetricCard
		title="Night Alerts"
		icon={mdiBellSleep}
		metric={metrics?.alerts.countNight || 0}
		comparison={{value: comparison?.alerts.countNight || 0}}
	/>
</div>
