<script lang="ts">
	import type { OncallRoster } from "$src/lib/api";
	import { Header } from "svelte-ux";

	type Props = { roster: OncallRoster };
	const { roster }: Props = $props();

	let metrics = $state({
		alertsLast24h: 3,
		averageResponseTime: "5m 12s",
		escalationsLast30d: 2,
		handoversCompleted: 12
	});
</script>

<Header title="Roster Stats" />
	
<div class="grid grid-cols-2 gap-4 mt-2">
	<div class="bg-surface-100 p-3 rounded-lg">
		<div class="text-sm text-surface-600">Alerts (Last 24h)</div>
		<div class="text-2xl font-semibold">{metrics.alertsLast24h}</div>
	</div>
	<div class="bg-surface-100 p-3 rounded-lg">
		<div class="text-sm text-surface-600">Avg. Response Time</div>
		<div class="text-2xl font-semibold">{metrics.averageResponseTime}</div>
	</div>
	<div class="bg-surface-100 p-3 rounded-lg">
		<div class="text-sm text-surface-600">Escalations (30d)</div>
		<div class="text-2xl font-semibold">{metrics.escalationsLast30d}</div>
	</div>
	<div class="bg-surface-100 p-3 rounded-lg">
		<div class="text-sm text-surface-600">Handovers Completed</div>
		<div class="text-2xl font-semibold">{metrics.handoversCompleted}</div>
	</div>
</div>