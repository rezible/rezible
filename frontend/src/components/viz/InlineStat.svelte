<script lang="ts" module>
	export type InlineStatProps = {
		title: string;
		subheading: string;
		value: any;
		comparison?: MetricComparison;
	}

</script>

<script lang="ts">
	import { Header } from "svelte-ux";
	import { comparisonClass, type MetricComparison } from "./utils";
	import ComparisonText from "./ComparisonText.svelte";

	const { title, subheading, value, comparison }: InlineStatProps = $props();
</script>

<Header {title} {subheading} class="p-2 px-4">
	<div class="ml-4 flex flex-col" slot="actions">
		<span class="text-2xl font-semibold self-end">{value}</span>
		{#if comparison}
			<ComparisonText {comparison} />
		{/if}
	</div>
</Header>