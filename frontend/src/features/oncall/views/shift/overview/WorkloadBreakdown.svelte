<script lang="ts">
	import { Card, Header, SpringValue } from "svelte-ux";
	import {
		BarChart,
		Bar,
		Tooltip,
		Legend,
		Chart,
		Svg,
		Arc,
		Text,
		LinearGradient,
		radiansToDegrees,
		Canvas,
	} from "layerchart";
	import { type OncallShift } from "$lib/api";
	import { formatDuration } from "date-fns";
	import { getHourLabel, isBusinessHours, type ShiftEvent } from "$features/oncall/lib/utils";
	import type { ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { cls } from "@layerstack/tailwind";
	import InlineStat, { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	import ChartWithStats from "$src/components/viz/ChartWithStats.svelte";

	type Props = {
		shiftEvents: ShiftEvent[];
		metrics: ShiftMetrics;
	};

	let { shiftEvents, metrics }: Props = $props();

	// const sleepDisruptionColor = $derived(getScoreColor(metrics.sleepDisruptionScore));
	// const workloadColor = $derived(getScoreColor(metrics.workloadScore));
	// const burdenColor = $derived(getScoreColor(metrics.burdenScore));

	const getScoreLabel = (score: number) => {
		if (score < 30) return "Low";
		if (score < 70) return "Moderate";
		return "High";
	};

	type HourEventDistribution = { hour: number; alerts: number; incidents: number };
	const hourlyDistribution = $derived.by(() => {
		const hours: HourEventDistribution[] = Array.from({ length: 24 }, (_, i) => ({
			hour: i,
			alerts: 0,
			incidents: 0,
		}));
		shiftEvents.forEach((ev) => {
			if (ev.eventType === "alert") hours[ev.timestamp.hour].alerts++;
			if (ev.eventType === "incident") hours[ev.timestamp.hour].incidents++;
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
	const peakHourLabel = $derived(peakAlertHours.map((v, hour) => getHourLabel(v.hour)).join(", "));
	const hourSegmentAngle = (2 * Math.PI) / 24;

	const alertHourArcBackgroundColor = (hour: number) => {
		if (isBusinessHours(hour)) return "fill-accent-900/10"
		if (hour > 5 && hour < 22) return "fill-warning-900/10";
		return "fill-danger-900/10";
	}
	const alertHourArcFillColor = (hour: number) => {
		const numAlerts = hourAlertCounts[hour];
		if (numAlerts === maxAlertCount) return "oklch(var(--color-danger))";
		if (isBusinessHours(hour)) return "oklch(var(--color-accent))";
		if (hour > 5 && hour < 22) return "oklch(var(--color-warning))"; // off-hours alert
		return "oklch(var(--color-warning))"; // night alert
	};

	const BurdenArcSegments = 50;
	const segmentAngle = (2 * Math.PI) / BurdenArcSegments;

	const burdenArcSegmentColor = (score: number) => {
		if (score > 0.7) return "fill-danger";
		if (score > 0.35) return "fill-warning";
		return "fill-success";
	};
	const burdenArcColor = $derived(burdenArcSegmentColor(metrics?.burdenScore ?? 0));

	const burdenStats = $derived<InlineStatProps[]>([
		{title: "High Severity Incidents", subheading: "Incidents with a severity of 1 or 2", value: metrics.totalIncidents},
		{title: "Sleep Disruption", subheading: `Based on ${metrics.nightAlerts} night alerts`, value: metrics.nightAlerts},
		{title: "KTLO Workload", subheading: `Based on backlog and ongoing incidents`, value: getScoreLabel(metrics.workloadScore)},
		{title: "Off-Hours Activity", subheading: `Time spent active outside of 8am-6pm`, value: formatDuration({ minutes: metrics.offHoursTime })},
	]);

	const alertStats = $derived<InlineStatProps[]>([
		{title: "Peak Alert Hour", subheading: `${maxAlertCount} alerts fired`, value: peakHourLabel},
		{title: "Stat 2", subheading: `desc`, value: ""},
		{title: "Stat 3", subheading: `desc`, value: ""},
		{title: "Stat 4", subheading: `desc`, value: ""},
	]);
</script>

<div class="flex flex-col gap-2 w-full p-2 border bg-surface-100/40 border-surface-content/10 rounded">
	<Header title="Burden Rating" subheading="Indicator of the human impact of this shift" />
	<ChartWithStats chart={burdenScoreCircle} stats={burdenStats} />
</div>

<div class="flex flex-col gap-2 w-full p-2 border bg-surface-100/40 border-surface-content/10 rounded">
	<Header title="Alerts" subheading="Alerts by time of day" class="" />
	<ChartWithStats chart={alertHoursCircle} stats={alertStats} reverse />
</div>

{#snippet burdenScoreCircle()}
	<div class="h-[250px] w-[250px] overflow-hidden grid place-self-center">
		<Chart>
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
		</Chart>
	</div>
{/snippet}

{#snippet alertHoursCircle()}
	<div class="h-[250px] w-[250px] overflow-hidden grid place-self-center">
		<Chart let:tooltip>
			<Canvas center>
				{#each { length: hourAlertCounts.length } as _, hour}
					{@const startAngle = hour * hourSegmentAngle}
					{@const endAngle = (hour + 1) * hourSegmentAngle}
					{@const numAlerts = hourAlertCounts[hour]}
					{@const fill = alertHourArcFillColor(hour)}
					{@const outerRadius = (30 + 70 * (numAlerts / maxAlertCount)) / 100}
					{@const tooltipLabel = `${getHourLabel(hour)} - ${numAlerts} alert${numAlerts > 1 ? "s" : ""}`}

					<Arc
						{startAngle}
						{endAngle}
						track
						outerRadius={1}
						class="{alertHourArcBackgroundColor(hour)} stroke-surface-content/20"
						onpointermove={(e) => tooltip?.show(e, tooltipLabel)}
						onpointerleave={() => tooltip?.hide()}
					/>

					{#if numAlerts > 0}
						<Arc
							{startAngle}
							{endAngle}
							track
							{fill}
							{outerRadius}
							class={cls(
								numAlerts &&
									"hover:scale-90 origin-center [transform-box:fill-box] transition-transform"
							)}
							onpointermove={(e) => tooltip?.show(e, tooltipLabel)}
							onpointerleave={() => tooltip?.hide()}
						/>
					{/if}
				{/each}
			</Canvas>
			<Tooltip.Root let:data>{data}</Tooltip.Root>
		</Chart>
	</div>
{/snippet}
