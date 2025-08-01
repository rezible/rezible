<script lang="ts">
	import { scaleOrdinal } from "d3-scale";
	import { getOncallRosterMetricsOptions } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { useOncallRosterViewState } from "$features/oncall-roster";
	
	const view = useOncallRosterViewState();

	const rosterId = $derived(view.rosterId);

	const metricsQuery = createQuery(() => getOncallRosterMetricsOptions({query: {rosterId}}));
	const metrics = $derived(metricsQuery.data?.data);

	const generateAlertsData = (days: number) => {
		const data = [];
		
		const now = new Date();
		for (let i = days - 1; i >= 0; i--) {
			const date = new Date();
			date.setDate(now.getDate() - i);
			
			// Generate random data with some pattern
			const business = Math.floor(Math.random() * 5);
			const outOfHours = Math.floor(Math.random() * 3);
			
			data.push({
				date,
				business,
				outOfHours
			});
		}
		
		return data;
	};

	const generateIncidentTypeData = (days: number) => {
		const multiplier = days / 7;
		return [
			{ value: 5 * multiplier, label: "Infrastructure", color: "#4ade80" },
			{ value: 8 * multiplier, label: "Application", color: "#60a5fa" },
			{ value: 3 * multiplier, label: "Database", color: "#f97316" },
			{ value: 2 * multiplier, label: "Network", color: "#8b5cf6" },
			{ value: 4 * multiplier, label: "Security", color: "#f43f5e" },
		];
	};

	let periodDays = $state(30);

	const alertsData = $derived(generateAlertsData(periodDays));
	const incidentTypesData = $derived(generateIncidentTypeData(periodDays));

	// Chart colors
	const alertColors = ["#4ade80", "#f87171"];
	const pieColors = $derived(scaleOrdinal()
		.domain(incidentTypesData.map(d => d.label))
		.range(incidentTypesData.map(d => d.color)));
</script>

<div class="overflow-y-auto flex flex-col h-full max-h-full min-h-0 flex-1">
	<div class="flex gap-4">
<pre>Workload distribution (across team members)
Oncall Burden score
Alert patterns (time of day, day of week)
Comparison with other rosters/company average</pre>
	</div>
</div>