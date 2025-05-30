<script lang="ts" module>
	import type { MetricComparison } from "./MetricComparisonLabel.svelte";
	export type InlineStatProps = {
		title: string;
		subheading: string;
		value: number | string;
		comparison?: MetricComparison;
	}

</script>

<script lang="ts">
	import Header from "$components/header/Header.svelte";
	import MetricComparisonLabel from "./MetricComparisonLabel.svelte";

	const { title, subheading, value, comparison }: InlineStatProps = $props();
</script>

<Header {title} {subheading} classes={{root: "p-2 px-4"}}>
	{#snippet actions()}
		<div class="ml-4 flex flex-col">
			<span class="text-2xl font-semibold self-end">{value}</span>
			{#if comparison}
				<MetricComparisonLabel {comparison} metricValue={value} />
			{/if}
		</div>
	{/snippet}
</Header>