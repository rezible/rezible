<script lang="ts">
	import { Icon } from "svelte-ux";
	import MetricComparisonLabel, { type MetricComparison } from "./MetricComparisonLabel.svelte";

	type Props = {
		title: string;
		icon: string;
		metric: number | string;
		comparison?: MetricComparison;
	}

	const { title, icon, metric, comparison }: Props = $props();
</script>

<div class="flex flex-col gap-1 border rounded py-3 px-4 border-neutral-content/10 bg-neutral-900/30 min-w-64">
	<div class="w-full flex justify-between items-center">
		<span class="">{title}</span>
		<span class=""><Icon data={icon} /></span>
	</div>
	<div class="w-full flex justify-between items-end">
		<span class="text-lg font-bold">{metric}</span>
		{#if comparison}
			<MetricComparisonLabel {comparison} metricValue={metric} />
		{/if}
	</div>
</div>