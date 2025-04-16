<script lang="ts">
	import type { OncallShiftMetrics } from "$lib/api";
	import { hour12, hour12Label } from "$lib/format.svelte";

	import { formatDuration } from "date-fns";
	import { Header } from "svelte-ux";

	import { isBusinessHours } from "$features/oncall/lib/utils";
	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";

	import * as echarts from "echarts";
	import EChart, { type ChartProps } from "$components/viz/echart/EChart.svelte";

	import { shiftViewStateCtx } from "../context.svelte";

	type Props = {
		metrics: OncallShiftMetrics;
	};
	let { metrics }: Props = $props();

	const viewState = shiftViewStateCtx.get();

	const getScoreLabel = (score: number) => {
		if (score < 30) return "Low";
		if (score < 70) return "Moderate";
		return "High";
	};

	const burdenStats = $derived<InlineStatProps[]>([
		{
			title: "High Severity Incidents",
			subheading: "Incidents with a severity of 1 or 2",
			value: metrics.incidents,
		},
		{
			title: "Sleep Disruption",
			subheading: `Based on ${metrics.nightAlerts} night alerts`,
			value: metrics.nightAlerts,
		},
		{
			title: "KTLO Workload",
			subheading: `Based on backlog and ongoing incidents`,
			value: getScoreLabel(metrics.workloadScore),
		},
		{
			title: "Off-Hours Activity",
			subheading: `Time spent active outside of 8am-6pm`,
			value: formatDuration({ minutes: metrics.offHoursActivityTime }, { zero: true }),
		},
	]);

	const burdenGaugeData = [
		{
			value: 20,
			name: "Metric 1",
			title: {
				offsetCenter: ["-75%", "85%"],
			},
			detail: {
				offsetCenter: ["-75%", "105%"],
			},
		},
		{
			value: 40,
			name: "Metric 2",
			title: {
				offsetCenter: ["0%", "85%"],
			},
			detail: {
				offsetCenter: ["0%", "105%"],
			},
		},
		{
			value: 35,
			name: "Metric 3",
			title: {
				offsetCenter: ["75%", "85%"],
			},
			detail: {
				offsetCenter: ["75%", "105%"],
			},
		},
	];

	const burdenChartOptions = $derived<ChartProps["options"]>({
		tooltip: {
			formatter: "{b} : {c}%",
		},
		series: [
			{
				type: "gauge",
				name: "Burden Score",
				anchor: { show: false },
				pointer: {
					icon: "path://M2.9,0.7L2.9,0.7c1.4,0,2.6,1.2,2.6,2.6v115c0,1.4-1.2,2.6-2.6,2.6l0,0c-1.4,0-2.6-1.2-2.6-2.6V3.3C0.3,1.9,1.4,0.7,2.9,0.7z",
					width: 3,
					length: "55%",
				},
				progress: { show: true, overlap: false, width: 6 },
				axisLine: { show: false },
				axisLabel: { color: "inherit" },
				data: burdenGaugeData,
				title: {
					fontSize: 14,
					color: "inherit",
				},
				detail: {
					width: 30,
					height: 10,
					fontSize: 12,
					backgroundColor: "inherit",
					borderRadius: 3,
					formatter: "{value}%",
				},
			},
		],
	});

	type HourEventDistribution = { hour: number; alerts: number; incidents: number };
	const hourlyDistribution = $derived.by(() => {
		const hours: HourEventDistribution[] = Array.from({ length: 24 }, (_, hour) => ({
			hour,
			alerts: 0,
			incidents: 0,
		}));
		viewState.filteredEvents.forEach(({attributes: a}) => {
			const hour = new Date(a.timestamp).getHours();
			if (a.kind === "alert") hours[hour].alerts++;
			if (a.kind === "incident") hours[hour].incidents++;
		});
		return hours;
	});
	const hourAlertCounts = $derived.by<number[]>(() => {
		const counts = new Array(24).fill(0);
		hourlyDistribution.forEach((d) => (counts[d.hour] += d.alerts));
		return counts;
	});
	const maxAlertCount = $derived(Math.max(...hourAlertCounts));
	const peakAlertHours = $derived(hourlyDistribution.filter((d) => d.alerts === maxAlertCount));
	const peakHourLabel = $derived(peakAlertHours.map((v, hour) => hour12Label(v.hour)).join(", "));

	const alertHourArcBackgroundColor = (hour: number) => {
		if (isBusinessHours(hour)) return "rgba(135, 206, 250, 0.2)";
		if (hour > 5 && hour < 22) return "rgba(225, 230, 170, 0.2)";
		return "rgba(70, 50, 120, 0.2)";
	};
	const alertHourArcFillColor = (hour: number) => {
		if (hourAlertCounts[hour] === maxAlertCount) return "rgba(210, 110, 140, 1)";
		if (isBusinessHours(hour)) return "rgba(100, 110, 120, 1)";
		if (hour > 5 && hour < 22) return "rgba(210, 210, 130, 0.6)"; // off-hours alert
		return "rgba(200, 190, 100, 0.8)"; // night alert
	};

	const alertStats = $derived<InlineStatProps[]>([
		{ title: "Peak Alert Hour", subheading: `${maxAlertCount} alerts fired`, value: peakHourLabel },
		// { title: "Stat 2", subheading: `desc`, value: "" },
		// { title: "Stat 3", subheading: `desc`, value: "" },
		// { title: "Stat 4", subheading: `desc`, value: "" },
	]);

	const alertHoursChartOptions = $derived<ChartProps["options"]>({
		series: [
			{
				name: "Background",
				type: "pie",
				radius: [10, 150],
				center: ["50%", "50%"],
				label: { show: false },
				emphasis: { scale: false },
				itemStyle: {
					color: ({ dataIndex: hour }) => alertHourArcBackgroundColor(hour),
					borderWidth: 1,
					borderColor: "rgba(180, 180, 180, 0.1)",
				},
				data: Array.from({ length: 24 }).map((_, hour) => ({ value: 1 })),
				z: 0,
			},
			{
				name: "Alerts in Hour",
				type: "pie",
				radius: [10, 150],
				center: ["50%", "50%"],
				roseType: "area",
				itemStyle: {
					color: ({ dataIndex: hour }) => alertHourArcFillColor(hour),
					borderRadius: 0,
				},
				label: { show: false },
				emphasis: {
					focus: "self",
					scale: false,
				},
				data: Array.from({ length: 24 }).map((_, hour) => {
					const alerts = hourAlertCounts[hour];
					return {
						value: alerts > 0 ? Math.round(30 + 70 * (alerts / maxAlertCount)) / 100 : 0,
						name: `${hour12(hour)}${hour >= 12 ? "PM" : "AM"}`,
					};
				}),
				z: 1,
			},
		],
		tooltip: {
			trigger: "item",
			formatter: (params) => {
				const hour = Array.isArray(params) ? params[0].dataIndex : params.dataIndex;
				const numAlerts = hourAlertCounts[hour];
				return `${hour12Label(hour)} - ${numAlerts} alert${numAlerts !== 1 ? "s" : ""}`;
			},
			confine: true,
			position: ["50%", "50%"],
		},
	});
</script>

<div class="flex flex-col gap-2 w-full p-2 border border-surface-content/10 rounded">
	<Header title="Burden Rating" subheading="Indicator of the human impact of this shift" />
	<ChartWithStats stats={burdenStats}>
		{#snippet chart()}
			<div class="h-[300px] w-[300px] overflow-hidden grid place-self-center">
				<EChart init={echarts.init} options={burdenChartOptions} />
			</div>
		{/snippet}
	</ChartWithStats>
</div>

<div class="flex flex-col gap-2 w-full p-2 border border-surface-content/10 rounded">
	<Header title="Alerts" subheading="Alerts by time of day" class="" />
	<ChartWithStats stats={alertStats} reverse>
		{#snippet chart()}
			<div class="h-[300px] w-[300px] overflow-hidden grid place-self-center">
				<EChart init={echarts.init} options={alertHoursChartOptions} />
			</div>
		{/snippet}
	</ChartWithStats>
</div>

{#snippet burdenScoreCircle()}
	<div class="h-[250px] w-[250px] overflow-hidden grid place-self-center">
		<!--Chart>
			<Canvas center>
				{#each { length: BurdenArcSegments } as _, segment}
					{@const pct = (segment / BurdenArcSegments) * 100}
					<Arc
						startAngle={segment * segmentAngle}
						endAngle={(segment + 1) * segmentAngle}
						innerRadius={-20}
						cornerRadius={4}
						padAngle={0.02}
						spring
						class={pct < metrics.burdenScore ? burdenArcColor : "fill-surface-content/10"}
						track={{ class: "fill-surface-content/10" }}
					/>
				{/each}

				<Text
					value={Math.round(metrics.burdenScore)}
					textAnchor="middle"
					verticalAnchor="middle"
					dy={16}
					class="text-6xl tabular-nums"
				/>
			</Canvas>
		</Chart-->
	</div>
{/snippet}
