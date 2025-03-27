<script lang="ts">
	import { Icon } from "svelte-ux";
	import { type MetricComparison } from "./utils";
	import ComparisonText from "./ComparisonText.svelte";

	type Props = {
		title: string;
		icon: string;
		metric: any;
		comparison?: MetricComparison;
	}

	const { title, icon, metric, comparison }: Props = $props();
</script>

<div class="flex flex-col border rounded py-3 px-4 border-neutral-content/10 bg-neutral-900/30 min-w-64">
	<div class="w-full flex justify-between items-center">
		<span class="">{title}</span>
		<span class=""><Icon data={icon} /></span>
	</div>
	<span class="text-lg font-bold">{metric}</span>
	{#if comparison}
		<ComparisonText {comparison} />
	{/if}
</div>