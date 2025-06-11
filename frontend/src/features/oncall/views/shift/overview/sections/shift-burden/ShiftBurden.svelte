<script lang="ts">
	import type { OncallShiftBurdenMetricWeights, OncallShiftMetrics, OncallShiftMetricsBurden } from "$lib/api";

	import Header from "$components/header/Header.svelte";
	import InlineStat, { type InlineStatProps } from "$components/viz/InlineStat.svelte";
	
	import SectionCard from "../SectionCard.svelte";
	import BurdenRadar from "./BurdenRadar.svelte";
	import MetricComparisonLabel from "$src/components/viz/MetricComparisonLabel.svelte";
	import { Collapse } from "svelte-ux";

	type Props = {
		metrics?: OncallShiftMetrics;
		weights?: OncallShiftBurdenMetricWeights;
	};
	const { metrics, weights }: Props = $props();

	const burdenStats = $derived([
		{
			title: "Event Frequency",
			subheading: `How often interruptions occur during your shift.`,
			value: 2.2,
			comparison: {value: 4.3},
		},
		{
			title: "Life Impact",
			subheading: `Disruption to personal time and sleep.`,
			value: 7.8,
			comparison: {value: 4.5},
		},
		{
			title: "Response Requirements",
			subheading: `Complexity and urgency of responses.`,
			value: 7.1,
			comparison: {value: 3.0},
		},
		{
			title: "Time Impact",
			subheading: `Total time spent actively working on operational toil.`,
			value: 6.4,
			comparison: {value: 4.2},
		},
		{
			title: "Isolation",
			subheading: `Availability of support and documentation available.`,
			value: 4.4,
			comparison: {value: 3.4},
		},
	]);
	
	const weight = .2;
	const burdenValue = $derived(burdenStats.reduce((prev, s) => (prev + (s.value * weight)), 0));
	const compValue = $derived(burdenStats.reduce((prev, s) => (prev + (s.comparison.value * weight)), 0));
</script>

<SectionCard>
	<Header title="Workload" subheading="Indicators of the human impact of this shift" />

	<div class="flex gap-4 items-center">
		<div class="w-1/3 grid place-items-center">
			<div class="h-[350px] w-[400px] overflow-hidden grid place-self-center">
				<BurdenRadar {burdenValue} {burdenStats} />
			</div>
		</div>

		<div class="flex flex-col w-full gap-2">
			<div class="border rounded-lg flex-1">
				<InlineStat 
					title="Burden Score" 
					subheading="Overall workload and stress level of the shift, derived from the below categories." 
					value={burdenValue.toFixed(2)}
					outOf="10"
					comparison={{value: compValue}} />
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
										<MetricComparisonLabel comparison={stat.comparison} metricValue={stat.value} />
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
	</div>
</SectionCard>
