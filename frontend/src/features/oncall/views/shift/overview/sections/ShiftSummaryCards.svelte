<script lang="ts">
	import { mdiBellAlert, mdiBellSleep, mdiClockTimeFive, mdiFire, mdiGauge } from "@mdi/js";

	import MetricCard from "$components/viz/MetricCard.svelte";
	import type { OncallShiftMetrics } from "$lib/api";

	type Props = {
		metrics?: OncallShiftMetrics;
		comparison?: OncallShiftMetrics;
	};

	let { metrics, comparison }: Props = $props();

</script>

<div class="grid grid-flow-col gap-1 overflow-hidden">
	<MetricCard
		title="Burden Score"
		icon={mdiGauge}
		format="raw"
		metric={metrics?.burden.finalScore || 0}
		comparison={{value: comparison?.burden.finalScore || 0}}
	/>
	<MetricCard
		title="Incidents"
		icon={mdiFire}
		metric={metrics?.incidents.total || 0}
		comparison={{value: comparison?.incidents.total || 0}}
	/>
	<MetricCard
		title="Time in Incidents"
		icon={mdiClockTimeFive}
		format="duration"
		metric={metrics?.incidents.responseTimeMinutes || 0}
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
