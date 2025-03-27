<script lang="ts">
	import { Button, Header, Icon, Menu, MenuItem, Toggle } from "svelte-ux";
	import { mdiBellAlert, mdiCalendar, mdiChartBar, mdiFire } from "@mdi/js";
	import { scaleOrdinal } from "d3-scale";
	import MetricCard from "$src/components/viz/MetricCard.svelte";
	import TimePeriodSelect from "$components/time-period-select/TimePeriodSelect.svelte";

	type Props = {};
	const {}: Props = $props();

	type RosterMetrics = {
		alerts: number;
		incidents: number;
		handoverCompletion: number;
		outOfHoursAlerts: number;
		oncallBurden: number;
		backlogBurnRate: number;
	};

	const metrics = $derived<RosterMetrics>({
		alerts: 3,
		incidents: 2,
		handoverCompletion: 0.9,
		outOfHoursAlerts: 8,
		oncallBurden: 64,
		backlogBurnRate: 1.1,
	});

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
	<Header title="Roster Stats" classes={{root: "gap-2 text-lg font-medium p-2"}}>
		<div slot="avatar">
			<Icon data={mdiChartBar} class="text-primary-300" />
		</div>
		
		<div class="justify-end" slot="actions">
			<TimePeriodSelect bind:selected={periodDays} />
		</div>
	</Header>

	<div class="flex gap-4">
	</div>
</div>