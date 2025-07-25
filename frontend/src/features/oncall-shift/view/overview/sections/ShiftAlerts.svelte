<script lang="ts">
	import type { OncallShiftMetrics } from "$lib/api";
	import { hour12, hour12Label } from "$lib/format.svelte";

	import { Button } from "svelte-ux";

	import { isBusinessHours } from "$features/oncall-shift/lib/utils";
	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";

	import * as echarts from "echarts";
	import EChart, { type ChartProps } from "$components/viz/echart/EChart.svelte";

	import { useShiftViewState } from "../../shiftViewState.svelte";
	import { mdiFilter } from "@mdi/js";
	import SectionCard from "./SectionCard.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {
		metrics?: OncallShiftMetrics;
	};
	let { metrics }: Props = $props();

	const viewState = useShiftViewState();

	let showFilters = $state(false);

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
	const peakHoursText = $derived(peakAlertHours.map((v, hour) => hour12Label(v.hour)).join(", "));
	const peakHourValueLabel = $derived(!!peakHoursText ? peakHoursText : "N/A");

	const alertHourArcBackgroundColor = (hour: number) => {
		if (isBusinessHours(hour)) return "rgba(135, 206, 250, 0.2)";
		if (hour > 5 && hour < 22) return "rgba(225, 230, 120, 0.2)";
		return "rgba(70, 50, 120, 0.2)";
	};
	const alertHourArcFillColor = (hour: number) => {
		const count = hourAlertCounts[hour];
		if (count === 0) return "rgba(100, 100, 100, .8)"
		if (count === maxAlertCount) return "rgba(140, 80, 10, 1)";
		if (isBusinessHours(hour)) return "rgba(100, 110, 120, 1)";
		if (hour > 5 && hour < 22) return "rgba(230, 230, 80, .6)"; // off-hours alert
		return "rgba(200, 190, 100, 0.8)"; // night alert
	};

	const stats = $derived<InlineStatProps[]>([
		{ title: "Peak Hours", subheading: `${maxAlertCount} alerts fired`, value: peakHourValueLabel },
		{ title: "Percent Annotated", subheading: `Number of alerts annotated with feedback`, value: "15%" },
		{ title: "Actionability", subheading: `Alerts that required human intervention`, value: "" },
		{ title: "Accuracy", subheading: `Alerts that correctly indicated an issue`, value: "" },
		{ title: "Documentation", subheading: `Alerts that had sufficient accompanying information`, value: "" },
	]);

	const MinRadius = 30;
	// TODO: maybe https://echarts.apache.org/examples/en/editor.html?c=bar-polar-stack-radial
	const alertHoursChartOptions = $derived<ChartProps["options"]>({
		series: [
			{
				name: "Background",
				type: "pie",
				radius: [15, 150],
				center: ["50%", "50%"],
				label: { show: false },
				emphasis: { scale: false },
				itemStyle: {
					color: ({ dataIndex: hour }) => alertHourArcBackgroundColor(hour),
					borderWidth: 1,
					borderColor: "rgba(180, 180, 180, 0.1)",
				},
				data: Array.from({ length: 24 }).map((_, hour) => ({ value: 0 })),
				z: 0,
			},
			{
				name: "Alerts in Hour",
				type: "pie",
				radius: [15, 150],
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

<SectionCard>
	<div class="h-fit flex flex-col gap-2">
		<Header title="Alerts" subheading="Alerts fired during this shift">
			{#snippet actions()}
				<Button icon={mdiFilter} iconOnly on:click={() => (showFilters = !showFilters)} />
			{/snippet}
		</Header>

		{#if showFilters}
			<div class="w-full h-12 border"></div>
		{/if}
	</div>

	<ChartWithStats {stats} reverse>
		{#snippet chart()}
			<div class="h-[300px] w-[300px] overflow-hidden grid place-self-center">
				<EChart init={echarts.init} options={alertHoursChartOptions} />
			</div>
		{/snippet}
	</ChartWithStats>
</SectionCard>