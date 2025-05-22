<script lang="ts">
	import type { OncallShiftMetrics } from "$lib/api";
	import { hour12, hour12Label } from "$lib/format.svelte";

	import { Header } from "svelte-ux";

	import { isBusinessHours } from "$features/oncall/lib/utils";
	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";

	import * as echarts from "echarts";
	import EChart, { type ChartProps } from "$components/viz/echart/EChart.svelte";

	import { shiftViewStateCtx } from "../../context.svelte";

	type Props = {
		metrics: OncallShiftMetrics;
	};
	let { metrics }: Props = $props();

	const viewState = shiftViewStateCtx.get();

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
	const peakAlertHours = $derived(hourlyDistribution.filter((d) => (d.alerts > 0 && d.alerts === maxAlertCount)));
	const peakHourLabel = $derived(peakAlertHours.map((v, hour) => hour12Label(v.hour)).join(", "));

	const alertHourArcBackgroundColor = (hour: number) => {
		if (isBusinessHours(hour)) return "rgba(135, 206, 250, 0.25)";
		if (hour > 5 && hour < 22) return "rgba(225, 230, 120, 0.25)";
		return "rgba(70, 50, 120, 0.25)";
	};
	const alertHourArcFillColor = (hour: number) => {
		const count = hourAlertCounts[hour];
		if (count === 0) return "rgba(100, 100, 100, .8)"
		if (count === maxAlertCount) return "rgba(210, 110, 140, 1)";
		if (isBusinessHours(hour)) return "rgba(100, 110, 120, 1)";
		if (hour > 5 && hour < 22) return "rgba(230, 230, 80, .6)"; // off-hours alert
		return "rgba(200, 190, 100, 0.8)"; // night alert
	};

	const alertStats = $derived<InlineStatProps[]>([
		{ title: "Peak Alert Hour", subheading: `${maxAlertCount} alerts fired`, value: peakHourLabel },
		// { title: "Stat 2", subheading: `desc`, value: "" },
		// { title: "Stat 3", subheading: `desc`, value: "" },
		// { title: "Stat 4", subheading: `desc`, value: "" },
	]);

	const MinRadius = 30;
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
						value: alerts > 0 ? Math.round(MinRadius + 70 * (alerts / maxAlertCount)) / 100 : 0,
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
	<Header title="Alerts" subheading="Alerts by time of day" class="" />
	<ChartWithStats stats={alertStats} reverse>
		{#snippet chart()}
			<div class="h-[300px] w-[300px] overflow-hidden grid place-self-center">
				<EChart init={echarts.init} options={alertHoursChartOptions} />
			</div>
		{/snippet}
	</ChartWithStats>
</div>
