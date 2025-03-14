<script lang="ts">
	import type { OncallRoster } from "$lib/api";
	import { Button, Header, Icon, Menu, MenuItem, Toggle } from "svelte-ux";
	import { mdiCalendar, mdiChartBar } from "@mdi/js";
	import { scaleLinear, scaleTime, scaleOrdinal } from "d3-scale";
	import { format } from "date-fns";
	import { Bar, Legend, Line, Area, Pie, Chart, BarChart, LineChart } from "layerchart";

	type Props = { roster: OncallRoster };
	const { roster }: Props = $props();

	// Time period filter options
	const timePeriods = [
		{ value: "7d", label: "Last 7 Days" },
		{ value: "30d", label: "Last 30 Days" },
		{ value: "90d", label: "Last 90 Days" },
	];
	let selectedTimePeriod = $state(timePeriods[1].value);

	// Mock data for metrics
	let metrics = $state({
		alertsLast24h: 3,
		averageResponseTime: "5m 12s",
		escalationsLast30d: 2,
		handoversCompleted: 12,
		outOfHoursAlerts: 8,
	});

	const generateAlertsData = (period: string) => {
		const days = period === "7d" ? 7 : period === "30d" ? 30 : 90;
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

	const generateIncidentTypeData = (period: string) => {
		// Adjust values based on selected period
		const multiplier = period === "7d" ? 1 : period === "30d" ? 4 : 12;
		return [
			{ value: 5 * multiplier, label: "Infrastructure", color: "#4ade80" },
			{ value: 8 * multiplier, label: "Application", color: "#60a5fa" },
			{ value: 3 * multiplier, label: "Database", color: "#f97316" },
			{ value: 2 * multiplier, label: "Network", color: "#8b5cf6" },
			{ value: 4 * multiplier, label: "Security", color: "#f43f5e" },
		];
	};

	const generateResponseTimeData = (period: string) => {
		const days = period === "7d" ? 7 : period === "30d" ? 30 : 90;
		const data = [];
		
		const now = new Date();
		for (let i = days - 1; i >= 0; i--) {
			const date = new Date();
			date.setDate(now.getDate() - i);
			
			// Generate random response times between 1-15 minutes
			const responseTime = Math.floor(Math.random() * 14) + 1;
			
			data.push({
				date,
				value: responseTime
			});
		}
		
		return data;
	};

	const alertsData = $derived(generateAlertsData(selectedTimePeriod));
	const incidentTypesData = $derived(generateIncidentTypeData(selectedTimePeriod));
	const responseTimeData = $derived(generateResponseTimeData(selectedTimePeriod));

	// Chart colors
	const alertColors = ["#4ade80", "#f87171"];
	const pieColors = $derived(scaleOrdinal()
		.domain(incidentTypesData.map(d => d.label))
		.range(incidentTypesData.map(d => d.color)));

	// Format functions
	const formatDate = (date: Date) => format(date, "MMM d");
	const formatMinutes = (value: number) => `${value} min`;
</script>

<Header title="Roster Stats" classes={{root: "gap-2 text-lg font-medium"}}>
	<div slot="avatar">
		<Icon data={mdiChartBar} class="text-primary-300" />
	</div>
	
	<div class="justify-end" slot="actions">
		<Toggle let:on={open} let:toggle let:toggleOff>
			<Button on:click={toggle} classes={{root: "flex gap-2 items-center"}}>
				{selectedTimePeriod}
				<Menu {open} on:close={toggleOff}>
					{#each timePeriods as period}
						<MenuItem on:click={() => (selectedTimePeriod = period.value)}>{period.label}</MenuItem>
					{/each}
				</Menu>
				<Icon data={mdiCalendar} />
			</Button>
		</Toggle>
	</div>
</Header>

<div class="grid grid-cols-2 gap-2">
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
	<div class="bg-surface-100 p-3 rounded-lg col-span-2">
		<div class="text-sm text-surface-600">Out of Hours Alerts</div>
		<div class="text-2xl font-semibold">{metrics.outOfHoursAlerts}</div>
	</div>
</div>

<div class="bg-surface-100 p-4 rounded-lg">
	<h3 class="text-lg font-medium mb-2 text-center">Average Response Time</h3>
	<div style="height: 300px;">
		<LineChart
			data={responseTimeData}
			x="date"
			y="value"
			renderContext="canvas"
			yDomain={[0, null]}
			xScale={scaleTime()}
			yScale={scaleLinear()}
			padding={{ top: 20, right: 20, bottom: 40, left: 40 }}
		/>
			<!-- <AxisY gridlines formatTick={formatMinutes} />
			<AxisX formatTick={formatDate} />
			<Line stroke="#60a5fa" strokeWidth={2} />
			<Area fill="#60a5fa" fillOpacity={0.2} /> -->
	</div>
</div>
