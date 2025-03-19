<script lang="ts">
  import { Card, Text } from "svelte-ux";
  import { getHourLabel } from "$features/oncall/lib/utils";
  import { useHourlyDistributionQuery } from "$features/oncall/lib/shift-queries";
  import type { OncallShift } from "$lib/api";
  import type { ShiftMetrics } from "$features/oncall/lib/utils";
  import { formatDuration, formatPercentage } from "$features/oncall/lib/utils";
  import { BarChart, XAxis, YAxis, Bar, Tooltip, Legend } from "layerchart";

  type Props = {
    shift: OncallShift;
    metrics: ShiftMetrics;
  };
  
  let { shift, metrics }: Props = $props();
  
  const hourlyDistributionQuery = useHourlyDistributionQuery(shift);
  
  $derived businessHoursPercentage = (metrics.businessHoursAlerts / metrics.totalAlerts) * 100 || 0;
  $derived offHoursPercentage = (metrics.offHoursAlerts / metrics.totalAlerts) * 100 || 0;
  
  $derived peakHourLabel = getHourLabel(metrics.peakAlertHour);
  
  $derived chartData = $hourlyDistributionQuery.data || [];
  $derived formattedChartData = chartData.map(item => ({
    ...item,
    hourLabel: getHourLabel(item.hour)
  }));
</script>

<Card class="p-4">
  <div class="flex items-center justify-between mb-4">
    <Text variant="title">Workload Distribution</Text>
    {#if $hourlyDistributionQuery.isLoading}
      <div class="text-sm text-gray-500">Loading...</div>
    {/if}
  </div>
  
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <div>
      <div class="mb-4">
        <Text variant="subtitle">Time Distribution</Text>
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
        <Text variant="subtitle">Peak Alert Time</Text>
        <div class="text-lg font-semibold">{peakHourLabel}</div>
      </div>
      
      <div>
        <Text variant="subtitle">Total Oncall Time</Text>
        <div class="text-lg font-semibold">{formatDuration(metrics.totalOncallTime)}</div>
      </div>
    </div>
    
    <div class="h-64">
      {#if !$hourlyDistributionQuery.isLoading && formattedChartData.length > 0}
        <BarChart data={formattedChartData} xKey="hourLabel">
          <XAxis />
          <YAxis />
          <Tooltip />
          <Legend />
          <Bar name="Alerts" dataKey="alerts" fill="#4f46e5" />
          <Bar name="Incidents" dataKey="incidents" fill="#ef4444" />
        </BarChart>
      {/if}
    </div>
  </div>
</Card>
