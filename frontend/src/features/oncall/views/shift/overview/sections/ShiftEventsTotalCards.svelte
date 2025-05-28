<script lang="ts">
	import { mdiBellAlert, mdiBellSleep, mdiFire } from "@mdi/js";

	import MetricCard from "$components/viz/MetricCard.svelte";
	import type { OncallShiftMetrics } from "$lib/api";

	type Props = {
		metrics?: OncallShiftMetrics;
		comparison?: OncallShiftMetrics;
	};

	let { metrics, comparison }: Props = $props();
</script>

<div class="grid grid-cols-3 gap-2">
	<MetricCard
		title="Incidents"
		icon={mdiFire}
		metric={metrics?.incidents.total || 0}
		comparison={{value: comparison?.incidents.total || 0}}
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
