<script lang="ts">
	import { Header } from "svelte-ux";
	import { formatDuration } from "date-fns";
	import type { OncallShiftMetrics } from "$lib/api";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";
	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";

	type Props = {
		metrics: OncallShiftMetrics;
		comparison: OncallShiftMetrics;
	};

	let { metrics, comparison }: Props = $props();

	const colors = [
		'oklch(var(--color-danger))',
		'oklch(var(--color-warning))',
		'oklch(var(--color-success))',
		'oklch(var(--color-info))',
	];

	const incidentSeries = $derived(metrics.incidentActivity?.map((v, i) => ({key: v.incidentId, value: v.minutes, color: colors[i % colors.length]})));
	const totalMinutes = $derived(metrics.incidentActivity?.reduce((prev, val) => (prev + val.minutes), 0));
	const totalTimeFormatted = $derived(formatDuration({minutes: totalMinutes}, {zero: true}));
	
	const incidentStats = $derived<InlineStatProps[]>([
		{title: "Alert to Incident Ratio", subheading: `Alerts that became incidents`, value: metrics.alertIncidentRate,
			comparison: {value: comparison.alertIncidentRate, positive: true}
		},
		// {title: "Stat 2", subheading: `desc`, value: 0},
		// {title: "Stat 3", subheading: `desc`, value: 0},
		// {title: "Stat 4", subheading: `desc`, value: 0},
	])
</script>

<div class="flex flex-col gap-2 w-full p-2 rounded border border-surface-content/10">
	<Header title="Incidents" subheading="" class="" />

	<ChartWithStats stats={incidentStats}>
		{#snippet chart()}
			<div class="h-[250px] w-[300px] overflow-auto">
			</div>
		{/snippet}
	</ChartWithStats>
</div>
