<script lang="ts">
	import type { Component } from "svelte";
	import { formatDuration } from "date-fns";
	import MetricComparisonLabel, { type MetricComparison } from "./MetricComparisonLabel.svelte";

	type Props = {
		title: string;
		icon: Component;
		metric: number | string;
		format?: "percentage" | "duration" | "raw";
		comparison?: MetricComparison;
	};

	const { title, icon: MetricIcon, metric, format = "percentage", comparison }: Props = $props();

	const formattedMetric = $derived.by(() => {
		if (typeof metric === "string" || format === "raw") return metric;
		if (format === "duration") return formatDuration({ hours: metric / 60 });
		return metric;
	});
</script>

<div
	class="flex flex-col gap-3 border rounded py-3 px-4 border-muted-foreground/10 bg-neutral-900/30 min-w-64"
>
	<div class="w-full flex justify-between gap-8 items-center">
		<span class="text-muted-foreground/60 leading-none">{title}</span>
		<span class=""><MetricIcon aria-hidden="true" /></span>
	</div>
	<div class="w-full flex gap-4 items-center justify-between">
		<span class="text-3xl font-bold">{formattedMetric}</span>
		{#if comparison && typeof metric === "number"}
			<MetricComparisonLabel {comparison} metricValue={metric} {format} />
		{/if}
	</div>
</div>
