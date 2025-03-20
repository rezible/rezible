<script lang="ts">
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import { Card, Header, SpringValue } from "svelte-ux";
	import { BarChart, Bar, Tooltip, Legend, Chart, Svg, Arc, Text, LinearGradient, radiansToDegrees } from "layerchart";
	import type { OncallShift } from "$lib/api";
	import { formatPercentage } from "$lib/format.svelte";
	import { formatDuration } from "date-fns";
	import { getHourLabel } from "$features/oncall/lib/utils";
	import type { ShiftMetrics } from "$features/oncall/lib/shift-metrics";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import { cls } from "@layerstack/tailwind";
	import Stat from "$src/components/viz/Stat.svelte";

	type Props = {
		shift: OncallShift;
		metrics?: ShiftMetrics;
	};

	let { shift, metrics }: Props = $props();

	function isBusinessHours(hour: number): boolean {
		return hour >= 9 && hour < 17;
	}

	const hourlyDistributionQuery = createQuery(() =>
		queryOptions({
			queryKey: ["hourlyDistribution", shift.id],
			queryFn: async () => {
				// Simulate API delay
				await new Promise((resolve) => setTimeout(resolve, 600));

				// Generate mock hourly distribution data
				const hours = Array.from({ length: 24 }, (_, i) => i);
				return hours.map((hour) => ({
					hour,
					alerts: Math.floor(Math.random() * (isBusinessHours(hour) ? 3 : 1.5)),
					incidents: Math.random() > 0.8 ? Math.floor(Math.random() * 2) : 0,
				}));
			},
			staleTime: 5 * 60 * 1000, // 5 minutes
		})
	);

	const chartData = $derived(hourlyDistributionQuery.data || []);
	const formattedChartData = $derived(
		chartData.map((item) => ({ ...item, hourLabel: getHourLabel(item.hour) }))
	);

	const BurdenArcSegments = 50;
	const segmentAngle = (2 * Math.PI) / BurdenArcSegments;

	function burdenColor(score: number) {
		if (score > .70) return "fill-danger";
		if (score > .35) return "fill-warning";
		return "fill-success";
	}
	const burdenArcColor = $derived(burdenColor(metrics?.burdenScore ?? 0));
</script>

{#snippet burdenScoreCircle(score: number)}
<div class="h-[240px] w-[240px] p-4 overflow-hidden">
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
</div>
{/snippet}

<Card title="Oncall Workload">
	<div class="flex flex-row flex-wrap gap-6 px-4">
		{#if !metrics}
			<LoadingIndicator />
		{:else}
			{@const businessHoursPercentage = (metrics.businessHoursAlerts / metrics.totalAlerts) * 100}
			{@const offHoursPercentage = (metrics.offHoursAlerts / metrics.totalAlerts) * 100}
			{@const peakHourLabel = getHourLabel(metrics.peakAlertHour)}

			<div class="border p-2">
				<Header title="Burden Rating" subheading="Indicator of the human impact of this shift" />
				{@render burdenScoreCircle(metrics.burdenScore)}
			</div>

			<div class="border p-2">
				<Stat 
					title="Time Outside of Business Hours" 
					value={formatDuration({ minutes: metrics.offHoursTime })} 
					description="Time spent active outside of 8am-6pm" />
			</div>

			<div class="border p-2">

				<div class="mb-4">
					<span>Time Distribution</span>
					<div class="flex justify-between mt-2">
						<div>
							<div class="text-lg font-semibold">
								{formatPercentage(businessHoursPercentage)}
							</div>
							<div class="text-sm text-gray-500">Business Hours</div>
						</div>
						<div>
							<div class="text-lg font-semibold">{formatPercentage(offHoursPercentage)}</div>
							<div class="text-sm text-gray-500">Off Hours</div>
						</div>
					</div>
				</div>

				<div class="mb-4">
					<span>Peak Alert Time</span>
					<div class="text-lg font-semibold">{peakHourLabel}</div>
				</div>
			</div>

			<div class="h-64">
				{#if !hourlyDistributionQuery.isLoading && formattedChartData.length > 0}
					chart
					<!--BarChart data={formattedChartData} xKey="hourLabel">
					<XAxis />
					<YAxis />
					<Tooltip />
					<Legend />
					<Bar name="Alerts" dataKey="alerts" fill="#4f46e5" />
					<Bar name="Incidents" dataKey="incidents" fill="#ef4444" />
				</BarChart-->
				{/if}
			</div>
		{/if}
	</div>
</Card>
