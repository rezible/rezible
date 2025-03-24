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
	} from "layerchart";
	import { type OncallShift } from "$lib/api";
	import { formatDuration } from "date-fns";
	import { getHourLabel, isBusinessHours, type ShiftEvent } from "$features/oncall/lib/utils";
	import type { ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import { cls } from "@layerstack/tailwind";

	type Props = {
		shift: OncallShift;
		shiftEvents: ShiftEvent[];
		metrics: ShiftMetrics;
	};

	let { shift, shiftEvents, metrics }: Props = $props();

	// const sleepDisruptionColor = $derived(getScoreColor(metrics.sleepDisruptionScore));
	// const workloadColor = $derived(getScoreColor(metrics.workloadScore));
	// const burdenColor = $derived(getScoreColor(metrics.burdenScore));

	const getScoreLabel = (score: number) => {
		if (score < 30) return "Low";
		if (score < 70) return "Moderate";
		return "High";
	};

	type HourEventDistribution = {hour: number, alerts: number, incidents: number};
	const hourlyDistribution = $derived.by(() => {
		const hours: HourEventDistribution[] = Array.from({ length: 24 }, (_, i) => ({hour: i, alerts: 0, incidents: 0}));
		shiftEvents.forEach(ev => {
			if (ev.eventType === "alert") hours[ev.timestamp.hour].alerts++;
			if (ev.eventType === "incident") hours[ev.timestamp.hour].incidents++;
		});
		return hours;
	});
	const hourAlertCounts = $derived.by<number[]>(() => {
		const counts = new Array(24).fill(0);
		hourlyDistribution.forEach(d => (counts[d.hour] += d.alerts));
		return counts;
	});
	const maxAlertCount = $derived(Math.max(...hourAlertCounts));
	const peakAlertHours = $derived(hourlyDistribution.filter((d) => d.alerts === maxAlertCount));
	const peakHourLabel = $derived(peakAlertHours.map((v, hour) => getHourLabel(v.hour)).join(", "));
	const hourSegmentAngle = (2 * Math.PI) / 24;

	const alertHourArcFillColor = (hour: number) => {
		const numAlerts = hourAlertCounts[hour];
		if (numAlerts === 0) return "fill-surface-content/10";
		if (numAlerts === maxAlertCount) return "oklch(var(--color-danger))"
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
</script>

{#snippet stat(title: string, subheading: string, value: string, comparison?: number)}
	<Header {title} {subheading} class="p-2 px-4">
		<div class="ml-4 flex flex-col" slot="actions">
			<span class="text-2xl font-semibold self-end">{value}</span>
			{#if comparison}
				<span class="text-xs text-surface-content self-end">{comparison}</span>
			{/if}
		</div>
	</Header>
{/snippet}

<div class="flex flex-col gap-2 w-full p-2 border bg-surface-100/40 border-surface-content/10 rounded">
	<Header title="Burden Rating" subheading="Indicator of the human impact of this shift" />

	<div class="grid grid-cols-3 gap-4">
		<div class="h-[250px] w-[250px] p-4 overflow-hidden place-self-center">
			{@render burdenScoreCircle(metrics.burdenScore)}
		</div>

		<div class="col-span-2 border rounded-lg flex flex-col divide-y">
			{@render stat(
				"High Severity Incidents",
				"Incidents with a severity of 1 or 2",
				metrics.totalIncidents.toString(),
			)}

			{@render stat(
				"Sleep Disruption",
				`Based on ${metrics.nightAlerts} night alerts`,
				metrics.nightAlerts.toString()
			)}

			{@render stat(
				"KTLO Workload",
				"Based on backlog and ongoing incidents",
				getScoreLabel(metrics.workloadScore)
			)}

			{@render stat(
				"Off-Hours Activity",
				"Time spent active outside of 8am-6pm",
				formatDuration({ minutes: metrics.offHoursTime })
			)}
		</div>
	</div>
</div>

<div class="flex flex-col gap-2 w-full p-2 border bg-surface-100/40 border-surface-content/10 rounded">
	<Header title="Alerts" subheading="Alerts by time of day" class="" />

	<div class="grid grid-cols-3 gap-4 items-center">
		<div class="col-span-2 border rounded flex flex-col divide-y h-fit">
			{@render stat("Peak Alert Hour", `${maxAlertCount} alerts fired`, peakHourLabel)}
			{@render stat("Stat 2", `desc`, "")}
			{@render stat("Stat 3", `desc`, "")}
			{@render stat("Stat 4", `desc`, "")}
		</div>

		<div class="h-[250px] w-[250px] border rounded-full overflow-hidden place-self-center">
			{@render alertHoursCircle()}
		</div>
	</div>
</div>

{#snippet burdenScoreCircle(score: number)}
	<Chart>
		<Svg center>
			<SpringValue value={score} let:value>
				{#each { length: BurdenArcSegments } as _, segment}
					{@const pct = (segment / BurdenArcSegments) * 100}
					<Arc
						startAngle={segment * segmentAngle}
						endAngle={(segment + 1) * segmentAngle}
						innerRadius={-20}
						cornerRadius={4}
						padAngle={0.02}
						spring
						class={pct < (value ?? 0) ? burdenArcColor : "fill-surface-content/10"}
						track={{ class: "fill-surface-content/10" }}
					/>
				{/each}

				<Text
					value={Math.round(value ?? 0)}
					textAnchor="middle"
					verticalAnchor="middle"
					dy={16}
					class="text-6xl tabular-nums"
				/>
			</SpringValue>
		</Svg>
	</Chart>
{/snippet}

{#snippet alertHoursCircle()}
	<Chart let:tooltip>
		<Svg center>
			{#each { length: hourAlertCounts.length } as _, hour}
				{@const startAngle = hour * hourSegmentAngle}
				{@const endAngle = (hour + 1) * hourSegmentAngle}
				{@const numAlerts = hourAlertCounts[hour]}
				{@const fill = alertHourArcFillColor(hour)}
				{@const outerRadius = (30 + 70 * (numAlerts / maxAlertCount)) / 100}
				{@const tooltipLabel = `${getHourLabel(hour)} - ${numAlerts} alert${numAlerts > 1 ? "s" : ""}`}

				{#if numAlerts !== maxAlertCount}
					<Arc
						{startAngle}
						{endAngle}
						track
						outerRadius={1}
						class="fill-surface-300/10 stroke-surface-content/10"
						onpointermove={(e) => numAlerts === 0 && tooltip?.show(e, tooltipLabel)}
						onpointerleave={() => tooltip?.hide()}
					/>
				{/if}

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
		</Svg>
		<Tooltip.Root let:data>{data}</Tooltip.Root>
	</Chart>
{/snippet}