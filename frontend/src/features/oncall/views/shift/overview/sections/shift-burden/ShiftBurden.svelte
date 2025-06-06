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
			subheading: `description`,
			value: 22,
			comparison: {value: 43},
		},
		{
			title: "Life Impact",
			subheading: `description`,
			value: 50,
			comparison: {value: 40},
		},
		{
			title: "Response Requirements",
			subheading: `description`,
			value: 40,
			comparison: {value: 40},
		},
		{
			title: "Time Impact",
			subheading: `description`,
			value: 20,
			comparison: {value: 40},
		},
		{
			title: "Isolation",
			subheading: `description`,
			value: 30,
			comparison: {value: 40},
		},
	]);
	
	const weight = .2;
	const score = $derived(burdenStats.reduce((prev, s) => (prev + (s.value * weight)), 0));
	const compScore = $derived(burdenStats.reduce((prev, s) => (prev + (s.comparison.value * weight)), 0));
</script>

<SectionCard>
	<Header title="Workload" subheading="Indications of the human impact of this shift" />

	<div class="flex gap-4 items-center">
		<div class="w-1/3 grid place-items-center">
			<div class="h-[350px] w-[400px] overflow-hidden grid place-self-center">
				<BurdenRadar {burdenStats} />
			</div>
		</div>

		<div class="flex flex-col w-full gap-2">
			<div class="border rounded-lg flex-1">
				<InlineStat title="Burden Score" subheading="" value={score} comparison={{value: compScore}} />
			</div>
			<div class="flex flex-col h-fit w-full place-self-start divide-y border rounded-lg">
				{#each burdenStats as stat}
					<Collapse classes={{icon: "mr-2"}}>
						<Header slot="trigger" title={stat.title} subheading={stat.subheading} classes={{root: "p-2 px-4 flex-1", title: "text-lg"}}>
							{#snippet actions()}
								<div class="ml-4 flex flex-col">
									<span class="text-lg font-semibold self-end">{stat.value}</span>
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
