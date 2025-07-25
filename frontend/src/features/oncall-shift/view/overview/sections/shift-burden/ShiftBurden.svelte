<script lang="ts">
	import { Collapse } from "svelte-ux";
	import type { OncallShiftMetrics } from "$lib/api";
	import MetricComparisonLabel from "$components/viz/MetricComparisonLabel.svelte";
	import Header from "$components/header/Header.svelte";
	
	import SectionCard from "../SectionCard.svelte";
	import BurdenRadar from "./BurdenRadar.svelte";

	type Props = {
		metrics?: OncallShiftMetrics;
		comparison?: OncallShiftMetrics;
	};
	const { metrics, comparison }: Props = $props();

	const burden = $derived(metrics?.burden);
	const comp = $derived(comparison?.burden);

	const burdenStats = $derived((!!burden && !!comp) ? [
		{
			title: "Event Frequency",
			subheading: `How often interruptions occur during your shift.`,
			value: burden.eventFrequency,
			comparison: {value: comp.eventFrequency},
		},
		{
			title: "Life Impact",
			subheading: `Disruption to personal time and sleep.`,
			value: burden.lifeImpact,
			comparison: {value: comp.lifeImpact},
		},
		{
			title: "Time Impact",
			subheading: `Total time spent actively working on operational toil.`,
			value: burden.timeImpact,
			comparison: {value: comp.timeImpact},
		},
		{
			title: "Response Requirements",
			subheading: `Complexity and urgency of responses.`,
			value: burden.responseRequirements,
			comparison: {value: comp.responseRequirements},
		},
		{
			title: "Isolation",
			subheading: `Availability of support and documentation available.`,
			value: burden.isolation,
			comparison: {value: comp.isolation},
		},
	] : []);
</script>

<SectionCard>
	<Header title="Workload" subheading="Indicators of the human impact of this shift" />

	<div class="flex gap-4 items-center">
		<div class="w-1/3 grid place-items-center">
			<div class="h-[350px] w-[400px] overflow-hidden grid place-self-center">
				{#if burdenStats.length > 0}
					<BurdenRadar {burdenStats} />
				{/if}
			</div>
		</div>

		<div class="flex flex-col h-fit w-full place-self-start divide-y border rounded-lg">
			{#each burdenStats as stat}
				<Collapse classes={{icon: "mr-2"}}>
					<Header slot="trigger" title={stat.title} subheading={stat.subheading} classes={{root: "p-2 px-4 flex-1", title: "text-lg"}}>
						{#snippet actions()}
							<div class="ml-4 flex flex-col">
								<span class="text-lg font-semibold self-end">
									{stat.value}
									<span class="text-xs text-surface-content/50">/ 10</span>
								</span>
								{#if stat.comparison}
									<MetricComparisonLabel comparison={stat.comparison} metricValue={stat.value} format="raw" />
								{/if}
							</div>
						{/snippet}
					</Header>
					<div class="p-3">
						TODO
					</div>
				</Collapse>
			{/each}
		</div>
	</div>
</SectionCard>
