<script lang="ts">
	import type { OncallRoster } from "$src/lib/api";
	import { Header, Select } from "svelte-ux";
	import * as echarts from 'echarts/core';
	import { BarChart, LineChart, PieChart } from 'echarts/charts';
	import {
		TitleComponent,
		TooltipComponent,
		GridComponent,
		DatasetComponent,
		TransformComponent,
		LegendComponent
	} from 'echarts/components';
	import { LabelLayout, UniversalTransition } from 'echarts/features';
	import { CanvasRenderer } from 'echarts/renderers';
	import EChart from "$src/components/echart/EChart.svelte";
	import { init } from "echarts/core";

	// Register ECharts components
	echarts.use([
		TitleComponent,
		TooltipComponent,
		GridComponent,
		DatasetComponent,
		TransformComponent,
		LegendComponent,
		BarChart,
		LineChart,
		PieChart,
		LabelLayout,
		UniversalTransition,
		CanvasRenderer
	]);

	type Props = { roster: OncallRoster };
	const { roster }: Props = $props();

	// Time period filter options
	const timePeriods = [
		{ value: '7d', label: 'Last 7 Days' },
		{ value: '30d', label: 'Last 30 Days' },
		{ value: '90d', label: 'Last 90 Days' }
	];
	let selectedTimePeriod = $state(timePeriods[1].value);

	// Mock data for metrics
	let metrics = $state({
		alertsLast24h: 3,
		averageResponseTime: "5m 12s",
		escalationsLast30d: 2,
		handoversCompleted: 12,
		outOfHoursAlerts: 8
	});

	// Mock data for charts
	const generateAlertsData = (period: string) => {
		const days = period === '7d' ? 7 : period === '30d' ? 30 : 90;
		const data = [];
		const businessHoursData = [];
		const outOfHoursData = [];
		
		const now = new Date();
		for (let i = days - 1; i >= 0; i--) {
			const date = new Date();
			date.setDate(now.getDate() - i);
			const dateStr = date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
			
			// Generate random data with some pattern
			const business = Math.floor(Math.random() * 5);
			const outOfHours = Math.floor(Math.random() * 3);
			
			data.push(dateStr);
			businessHoursData.push(business);
			outOfHoursData.push(outOfHours);
		}
		
		return {
			dates: data,
			business: businessHoursData,
			outOfHours: outOfHoursData
		};
	};

	const generateIncidentTypeData = (period: string) => {
		// Adjust values based on selected period
		const multiplier = period === '7d' ? 1 : period === '30d' ? 4 : 12;
		return [
			{ value: 5 * multiplier, name: 'Infrastructure' },
			{ value: 8 * multiplier, name: 'Application' },
			{ value: 3 * multiplier, name: 'Database' },
			{ value: 2 * multiplier, name: 'Network' },
			{ value: 4 * multiplier, name: 'Security' }
		];
	};

	const generateResponseTimeData = (period: string) => {
		const days = period === '7d' ? 7 : period === '30d' ? 30 : 90;
		const data = [];
		const responseTimeData = [];
		
		const now = new Date();
		for (let i = days - 1; i >= 0; i--) {
			const date = new Date();
			date.setDate(now.getDate() - i);
			const dateStr = date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
			
			// Generate random response times between 1-15 minutes
			const responseTime = Math.floor(Math.random() * 14) + 1;
			
			data.push(dateStr);
			responseTimeData.push(responseTime);
		}
		
		return {
			dates: data,
			responseTimes: responseTimeData
		};
	};

	// Chart options
	$effect(() => {
		alertsChartOptions = getAlertsChartOptions(selectedTimePeriod);
		incidentTypesChartOptions = getIncidentTypesChartOptions(selectedTimePeriod);
		responseTimeChartOptions = getResponseTimeChartOptions(selectedTimePeriod);
	});

	let alertsChartOptions = $state({});
	let incidentTypesChartOptions = $state({});
	let responseTimeChartOptions = $state({});

	const getAlertsChartOptions = (period: string) => {
		const alertsData = generateAlertsData(period);
		return {
			title: {
				text: 'Alerts Over Time',
				left: 'center'
			},
			tooltip: {
				trigger: 'axis',
				axisPointer: {
					type: 'shadow'
				}
			},
			legend: {
				data: ['Business Hours', 'Out of Hours'],
				bottom: 0
			},
			xAxis: {
				type: 'category',
				data: alertsData.dates
			},
			yAxis: {
				type: 'value',
				name: 'Number of Alerts'
			},
			series: [
				{
					name: 'Business Hours',
					type: 'bar',
					stack: 'total',
					data: alertsData.business,
					color: '#4ade80'
				},
				{
					name: 'Out of Hours',
					type: 'bar',
					stack: 'total',
					data: alertsData.outOfHours,
					color: '#f87171'
				}
			]
		};
	};

	const getIncidentTypesChartOptions = (period: string) => {
		const incidentTypeData = generateIncidentTypeData(period);
		return {
			title: {
				text: 'Incidents by Type',
				left: 'center'
			},
			tooltip: {
				trigger: 'item',
				formatter: '{a} <br/>{b}: {c} ({d}%)'
			},
			legend: {
				orient: 'horizontal',
				bottom: 0
			},
			series: [
				{
					name: 'Incident Types',
					type: 'pie',
					radius: ['40%', '70%'],
					avoidLabelOverlap: false,
					itemStyle: {
						borderRadius: 10,
						borderColor: '#fff',
						borderWidth: 2
					},
					label: {
						show: false,
						position: 'center'
					},
					emphasis: {
						label: {
							show: true,
							fontSize: 16,
							fontWeight: 'bold'
						}
					},
					labelLine: {
						show: false
					},
					data: incidentTypeData
				}
			]
		};
	};

	const getResponseTimeChartOptions = (period: string) => {
		const responseTimeData = generateResponseTimeData(period);
		return {
			title: {
				text: 'Average Response Time',
				left: 'center'
			},
			tooltip: {
				trigger: 'axis',
				formatter: function(params: any) {
					return `${params[0].name}: ${params[0].value} minutes`;
				}
			},
			xAxis: {
				type: 'category',
				data: responseTimeData.dates
			},
			yAxis: {
				type: 'value',
				name: 'Minutes'
			},
			series: [
				{
					data: responseTimeData.responseTimes,
					type: 'line',
					smooth: true,
					color: '#60a5fa',
					areaStyle: {
						color: {
							type: 'linear',
							x: 0,
							y: 0,
							x2: 0,
							y2: 1,
							colorStops: [
								{
									offset: 0,
									color: 'rgba(96, 165, 250, 0.5)'
								},
								{
									offset: 1,
									color: 'rgba(96, 165, 250, 0.05)'
								}
							]
						}
					}
				}
			]
		};
	};
</script>

<Header title="Roster Stats" />

<div class="flex justify-end mb-4">
	<Select 
		bind:value={selectedTimePeriod}
		options={timePeriods}
		class="w-40"
	/>
</div>
	
<div class="grid grid-cols-2 gap-4 mb-6">
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

<div class="grid grid-cols-2 gap-6 mb-6">
	<div class="bg-surface-100 p-4 rounded-lg">
		<div style="height: 300px;">
			<EChart init={init} options={alertsChartOptions} />
		</div>
	</div>
	<div class="bg-surface-100 p-4 rounded-lg">
		<div style="height: 300px;">
			<EChart init={init} options={incidentTypesChartOptions} />
		</div>
	</div>
</div>

<div class="bg-surface-100 p-4 rounded-lg mb-6">
	<div style="height: 300px;">
		<EChart init={init} options={responseTimeChartOptions} />
	</div>
</div>
