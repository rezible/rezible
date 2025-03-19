<script lang="ts">
  import { Card, Text, ProgressBar } from "svelte-ux";
  import type { ShiftMetrics } from "$features/oncall/lib/utils";

  type Props = {
    metrics: ShiftMetrics;
    loading: boolean;
  };
  
  let { metrics, loading }: Props = $props();
  
  $derived getScoreColor = (score: number) => {
    if (score < 30) return "bg-green-500";
    if (score < 70) return "bg-yellow-500";
    return "bg-red-500";
  };
  
  $derived sleepDisruptionColor = getScoreColor(metrics.sleepDisruptionScore);
  $derived workloadColor = getScoreColor(metrics.workloadScore);
  $derived burdenColor = getScoreColor(metrics.burdenScore);
  
  $derived getScoreLabel = (score: number) => {
    if (score < 30) return "Low";
    if (score < 70) return "Moderate";
    return "High";
  };
</script>

<Card class="p-4">
  <div class="flex items-center justify-between mb-4">
    <Text variant="title">Health Indicators</Text>
    {#if loading}
      <div class="text-sm text-gray-500">Loading...</div>
    {/if}
  </div>
  
  <div class="space-y-6">
    <div>
      <div class="flex justify-between mb-1">
        <Text variant="subtitle">Sleep Disruption</Text>
        <span class="font-medium">{getScoreLabel(metrics.sleepDisruptionScore)}</span>
      </div>
      <ProgressBar value={metrics.sleepDisruptionScore} max={100} class={sleepDisruptionColor} />
      <div class="text-sm text-gray-500 mt-1">
        Based on {metrics.nightAlerts} night alerts
      </div>
    </div>
    
    <div>
      <div class="flex justify-between mb-1">
        <Text variant="subtitle">Workload</Text>
        <span class="font-medium">{getScoreLabel(metrics.workloadScore)}</span>
      </div>
      <ProgressBar value={metrics.workloadScore} max={100} class={workloadColor} />
      <div class="text-sm text-gray-500 mt-1">
        Compared to team average
      </div>
    </div>
    
    <div>
      <div class="flex justify-between mb-1">
        <Text variant="subtitle">Overall Burden</Text>
        <span class="font-medium">{getScoreLabel(metrics.burdenScore)}</span>
      </div>
      <ProgressBar value={metrics.burdenScore} max={100} class={burdenColor} />
      <div class="text-sm text-gray-500 mt-1">
        Aggregate score based on all metrics
      </div>
    </div>
  </div>
</Card>
