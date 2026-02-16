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
		metric={metrics?.events.totalIncidents || 0}
		comparison={{value: comparison?.events.totalIncidents || 0}}
	/>
	<MetricCard
		title="Time Responding to Interrupts"
		icon={mdiClockTimeFive}
		format="duration"
		metric={metrics?.events.interruptResponseTime || 0}
		comparison={{value: comparison?.events.interruptResponseTime || 0}}
	/>
	<MetricCard
		title="Alerts"
		icon={mdiBellAlert}
		metric={metrics?.events.totalAlerts || 0}
		comparison={{value: comparison?.events.totalAlerts || 0}}
	/>
	<MetricCard
		title="Night Interrupts"
		icon={mdiBellSleep}
		metric={metrics?.events.interruptsNight || 0}
		comparison={{value: comparison?.events.interruptsNight || 0}}
	/>
</div>
