<script lang="ts" module>
	import type { MetricComparison } from "./MetricComparisonLabel.svelte";
	export type InlineStatProps = {
		title: string;
		subheading: string;
		value: number | string;
		outOf?: string;
		small?: boolean;
		format?: "percentage" | "duration" | "raw";
		comparison?: MetricComparison;
	}

</script>

<script lang="ts">
	import Header from "$components/header/Header.svelte";
	import MetricComparisonLabel from "./MetricComparisonLabel.svelte";

	const { title, subheading, value, outOf, small = false, format = "percentage", comparison }: InlineStatProps = $props();

	const valueText = $derived(typeof value === "string" ? value : value.toFixed(2));
</script>

<Header {title} {subheading} classes={{root: "p-2 px-4", title: (small ? "text-lg" : "text-xl")}}>
	{#snippet actions()}
		<div class="flex flex-col">
			<span class="{small ? "text-lg" : "text-2xl"} font-semibold self-end">
				{valueText}
				{#if outOf}
					<span class="text-xs font-bold text-surface-content/50">/ {outOf}</span>
				{/if}
			</span>
			{#if comparison && typeof value === "number"}
				<MetricComparisonLabel {comparison} metricValue={value} {format} />
			{/if}
		</div>
	{/snippet}
</Header>