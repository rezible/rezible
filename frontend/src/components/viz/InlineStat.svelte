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
	import { Header } from "svelte-ux";
	import ComparisonText from "./MetricComparisonLabel.svelte";

	const { title, subheading, value, comparison }: InlineStatProps = $props();
</script>

<Header {title} {subheading} class="p-2 px-4">
	<div class="ml-4 flex flex-col" slot="actions">
		<span class="text-2xl font-semibold self-end">{value}</span>
		{#if comparison}
			<ComparisonText {comparison} metricValue={value} />
		{/if}
	</div>
</Header>