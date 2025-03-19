<script lang="ts">
  import { Card, Text } from "svelte-ux";
  import type { ComparisonMetrics, ShiftMetrics } from "$features/oncall/lib/utils";
  import { formatComparisonValue } from "$features/oncall/lib/utils";
  import { PieChart, Pie, Cell, Legend, Tooltip } from "layerchart";

  type Props = {
    metrics: ShiftMetrics;
    comparison: ComparisonMetrics;
    loading: boolean;
  };
  
  let { metrics, comparison, loading }: Props = $props();
  
  $derived comparisonClass = (value: number) => {
    if (value > 0) return "text-red-500";
    if (value < 0) return "text-green-500";
    return "text-gray-500";
  };
  
  const COLORS = {
    critical: "#ef4444", // red
    high: "#f97316",     // orange
    medium: "#eab308",   // yellow
    low: "#22c55e"       // green
  };
  
  $derived pieData = [
    { name: "Critical", value: metrics.severityBreakdown.critical, color: COLORS.critical },
    { name: "High", value: metrics.severityBreakdown.high, color: COLORS.high },
    { name: "Medium", value: metrics.severityBreakdown.medium, color: COLORS.medium },
    { name: "Low", value: metrics.severityBreakdown.low, color: COLORS.low }
  ].filter(item => item.value > 0);
  
  $derived hasSeverityData = pieData.length > 0;
</script>

<Card class="p-4">
  <div class="flex items-center justify-between mb-4">
    <Text variant="title">Severity Breakdown</Text>
    {#if loading}
      <div class="text-sm text-gray-500">Loading...</div>
    {/if}
  </div>
  
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <div class="h-64">
      {#if hasSeverityData && !loading}
        <PieChart>
          <Pie
            data={pieData}
            dataKey="value"
            nameKey="name"
            cx="50%"
            cy="50%"
            outerRadius={80}
            label
          >
            {#each pieData as entry, index}
              <Cell key={`cell-${index}`} fill={entry.color} />
            {/each}
          </Pie>
          <Tooltip />
          <Legend />
        </PieChart>
      {:else if !loading}
        <div class="flex h-full items-center justify-center text-gray-500">
          No severity data available
        </div>
      {/if}
    </div>
    
    <div>
      <Text variant="subtitle">Comparison to Roster Average</Text>
      <div class="space-y-3 mt-3">
        {#if metrics.severityBreakdown.critical > 0}
          <div class="flex justify-between items-center">
            <div class="flex items-center">
              <div class="w-3 h-3 rounded-full mr-2" style="background-color: {COLORS.critical}"></div>
              <span>Critical</span>
            </div>
            <div class={comparisonClass(comparison.severityComparison.critical)}>
              {formatComparisonValue(comparison.severityComparison.critical)}
            </div>
          </div>
        {/if}
        
        {#if metrics.severityBreakdown.high > 0}
          <div class="flex justify-between items-center">
            <div class="flex items-center">
              <div class="w-3 h-3 rounded-full mr-2" style="background-color: {COLORS.high}"></div>
              <span>High</span>
            </div>
            <div class={comparisonClass(comparison.severityComparison.high)}>
              {formatComparisonValue(comparison.severityComparison.high)}
            </div>
          </div>
        {/if}
        
        {#if metrics.severityBreakdown.medium > 0}
          <div class="flex justify-between items-center">
            <div class="flex items-center">
              <div class="w-3 h-3 rounded-full mr-2" style="background-color: {COLORS.medium}"></div>
              <span>Medium</span>
            </div>
            <div class={comparisonClass(comparison.severityComparison.medium)}>
              {formatComparisonValue(comparison.severityComparison.medium)}
            </div>
          </div>
        {/if}
        
        {#if metrics.severityBreakdown.low > 0}
          <div class="flex justify-between items-center">
            <div class="flex items-center">
              <div class="w-3 h-3 rounded-full mr-2" style="background-color: {COLORS.low}"></div>
              <span>Low</span>
            </div>
            <div class={comparisonClass(comparison.severityComparison.low)}>
              {formatComparisonValue(comparison.severityComparison.low)}
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</Card>
