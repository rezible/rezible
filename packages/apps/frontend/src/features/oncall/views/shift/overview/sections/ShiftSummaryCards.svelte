<script lang="ts">
	import RiNotification2Line from "remixicon-svelte/icons/notification-2-line";
	import RiNotificationOffLine from "remixicon-svelte/icons/notification-off-line";
	import RiTimeLine from "remixicon-svelte/icons/time-line";
	import RiFireLine from "remixicon-svelte/icons/fire-line";
	import RiDashboardLine from "remixicon-svelte/icons/dashboard-line";

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
		icon={RiDashboardLine}
		format="raw"
		metric={metrics?.burden.finalScore || 0}
		comparison={{ value: comparison?.burden.finalScore || 0 }}
	/>
	<MetricCard
		title="Incidents"
		icon={RiFireLine}
		metric={metrics?.events.totalIncidents || 0}
		comparison={{ value: comparison?.events.totalIncidents || 0 }}
	/>
	<MetricCard
		title="Time Responding to Interrupts"
		icon={RiTimeLine}
		format="duration"
		metric={metrics?.events.interruptResponseTime || 0}
		comparison={{ value: comparison?.events.interruptResponseTime || 0 }}
	/>
	<MetricCard
		title="Alerts"
		icon={RiNotification2Line}
		metric={metrics?.events.totalAlerts || 0}
		comparison={{ value: comparison?.events.totalAlerts || 0 }}
	/>
	<MetricCard
		title="Night Interrupts"
		icon={RiNotificationOffLine}
		metric={metrics?.events.interruptsNight || 0}
		comparison={{ value: comparison?.events.interruptsNight || 0 }}
	/>
</div>
