<script lang="ts">
	import type { OncallShiftMetrics } from "$lib/api";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";
	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	import { mdiFilter } from "@mdi/js";
	import SectionCard from "./SectionCard.svelte";
	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";

	type Props = {
		metrics?: OncallShiftMetrics;
		comparison?: OncallShiftMetrics;
	};

	let { metrics, comparison }: Props = $props();

	let showFilters = $state(false);

	const colors = [
		'oklch(var(--color-danger))',
		'oklch(var(--color-warning))',
		'oklch(var(--color-success))',
		'oklch(var(--color-info))',
	];

	// const incidentSeries = $derived(metrics.incidentActivity?.map((v, i) => ({key: v.incidentId, value: v.minutes, color: colors[i % colors.length]})));
	// const totalMinutes = $derived(metrics.incidentActivity?.reduce((prev, val) => (prev + val.minutes), 0));
	// const totalTimeFormatted = $derived(formatDuration({minutes: totalMinutes}, {zero: true}));

	const stats = $derived<InlineStatProps[]>([
		{
			title: "Average Incident Severity",
			subheading: `Highest severity is 0`,
			value: 0,
			comparison: {value: 0, positive: true}
		},
		{
			title: "Alert to Incident Rate",
			subheading: `Alerts that became incidents`,
			value: metrics?.events.alertIncidentRate || 0,
			comparison: {value: comparison?.events.alertIncidentRate || 0, positive: true}
		},
		{
			title: "Incidents Reviewed",
			subheading: `From incident review meetings`,
			value: 0,
			comparison: {value: 0, positive: true}
		},
	])
</script>

<SectionCard>
	<div class="h-fit flex flex-col gap-2">
		<Header title="Incidents" subheading="Incidents opened during shift">
			{#snippet actions()}
				<Button onclick={() => (showFilters = !showFilters)}>filter</Button>
			{/snippet}
		</Header>

		{#if showFilters}
			<div class="w-full h-12 border"></div>
		{/if}
	</div>

	<ChartWithStats {stats}>
		{#snippet chart()}
			<div class="h-[250px] w-[300px] overflow-auto">
				
			</div>
		{/snippet}
	</ChartWithStats>
</SectionCard>
