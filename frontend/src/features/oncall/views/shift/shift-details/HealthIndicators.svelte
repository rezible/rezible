<script lang="ts">
	import { Card } from "svelte-ux";
	import type { ShiftMetrics } from "$features/oncall/lib/utils";

	type Props = {
		metrics: ShiftMetrics;
		loading: boolean;
	};

	let { metrics, loading }: Props = $props();

	const getScoreColor = (score: number) => {
		if (score < 30) return "bg-green-500";
		if (score < 70) return "bg-yellow-500";
		return "bg-red-500";
	};

	const sleepDisruptionColor = $derived(getScoreColor(metrics.sleepDisruptionScore));
	const workloadColor = $derived(getScoreColor(metrics.workloadScore));
	const burdenColor = $derived(getScoreColor(metrics.burdenScore));

	const getScoreLabel = (score: number) => {
		if (score < 30) return "Low";
		if (score < 70) return "Moderate";
		return "High";
	};
</script>

<Card class="p-4">
	<div class="flex items-center justify-between mb-4">
		<span>Health Indicators</span>
		{#if loading}
			<div class="text-sm text-gray-500">Loading...</div>
		{/if}
	</div>

	<div class="space-y-6">
		<div>
			<div class="flex justify-between mb-1">
				<span>Sleep Disruption</span>
				<span class="font-medium">{getScoreLabel(metrics.sleepDisruptionScore)}</span>
			</div>
			<!-- <ProgressBar value={metrics.sleepDisruptionScore} max={100} class={sleepDisruptionColor} /> -->
			<div class="text-sm text-gray-500 mt-1">
				Based on {metrics.nightAlerts} night alerts
			</div>
		</div>

		<div>
			<div class="flex justify-between mb-1">
				<span>Workload</span>
				<span class="font-medium">{getScoreLabel(metrics.workloadScore)}</span>
			</div>
			<!-- <ProgressBar value={metrics.workloadScore} max={100} class={workloadColor} /> -->
			<div class="text-sm text-gray-500 mt-1">Compared to team average</div>
		</div>

		<div>
			<div class="flex justify-between mb-1">
				<span>Overall Burden</span>
				<span class="font-medium">{getScoreLabel(metrics.burdenScore)}</span>
			</div>
			<!-- <ProgressBar value={metrics.burdenScore} max={100} class={burdenColor} /> -->
			<div class="text-sm text-gray-500 mt-1">Aggregate score based on all metrics</div>
		</div>
	</div>
</Card>
