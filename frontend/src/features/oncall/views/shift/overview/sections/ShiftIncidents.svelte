<script lang="ts">
	import { Button } from "svelte-ux";
	import type { OncallShiftMetrics } from "$lib/api";
	import ChartWithStats from "$components/viz/ChartWithStats.svelte";
	import { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	import { mdiFilter } from "@mdi/js";
	import SectionCard from "./SectionCard.svelte";
	import Header from "$src/components/header/Header.svelte";

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

	const incidentStats = $derived<InlineStatProps[]>([
		{
			title: "Alert to Incident Ratio",
			subheading: `Alerts that became incidents`,
			value: metrics?.alerts.incidentRate || 0,
			comparison: {value: comparison?.alerts.incidentRate || 0, positive: true}
		},
		{title: "Incidents by Severity", subheading: `TODO`, value: 0},
		// {title: "Stat 4", subheading: `desc`, value: 0},
	])
</script>

<SectionCard>
	<div class="h-fit flex flex-col gap-2">
		<Header title="Incidents" subheading="Incidents opened during shift">
			{#snippet actions()}
				<Button icon={mdiFilter} iconOnly on:click={() => (showFilters = !showFilters)} />
			{/snippet}
		</Header>

		{#if showFilters}
			<div class="w-full h-12 border"></div>
		{/if}
	</div>

	<ChartWithStats stats={incidentStats}>
		{#snippet chart()}
			<div class="h-[250px] w-[300px] overflow-auto">
			</div>
		{/snippet}
	</ChartWithStats>
</SectionCard>
