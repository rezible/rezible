<script lang="ts">
	import { Card } from "svelte-ux";
	import { getHourLabel } from "$features/oncall/lib/utils";
	import type { OncallShift } from "$lib/api";
	import type { ShiftMetrics } from "$features/oncall/lib/utils";
	import { formatDuration, formatPercentage } from "$features/oncall/lib/utils";
	import { BarChart, Bar, Tooltip, Legend } from "layerchart";
	import { createQuery, queryOptions } from "@tanstack/svelte-query";

	type Props = {
		shift: OncallShift;
		metrics: ShiftMetrics;
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

	const businessHoursPercentage = $derived((metrics.businessHoursAlerts / metrics.totalAlerts) * 100 || 0);
	const offHoursPercentage = $derived((metrics.offHoursAlerts / metrics.totalAlerts) * 100 || 0);

	const peakHourLabel = $derived(getHourLabel(metrics.peakAlertHour));

	const chartData = $derived(hourlyDistributionQuery.data || []);
	const formattedChartData = $derived(
		chartData.map((item) => ({ ...item, hourLabel: getHourLabel(item.hour) }))
	);
</script>

<Card class="p-4" title="Workload Distribution">
	<div class="flex items-center justify-between mb-4">
		{#if hourlyDistributionQuery.isLoading}
			<div class="text-sm text-gray-500">Loading...</div>
		{/if}
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
		<div>
			<div class="mb-4">
				<span>Time Distribution</span>
				<div class="flex justify-between mt-2">
					<div>
						<div class="text-lg font-semibold">{formatPercentage(businessHoursPercentage)}</div>
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

			<div>
				<span>Total Oncall Time</span>
				<div class="text-lg font-semibold">{formatDuration(metrics.totalOncallTime)}</div>
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
	</div>
</Card>
