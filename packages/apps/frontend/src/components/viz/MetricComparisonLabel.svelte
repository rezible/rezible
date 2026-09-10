<script lang="ts" module>
	export type MetricComparison = {
		value: number;
		averageMargin?: number;
		positive?: boolean;
		deltaLabel?: string;
		hint?: string;
	};
</script>

<script lang="ts">
	import type { Component } from "svelte";
	import RiArrowRightDownLine from "remixicon-svelte/icons/arrow-right-down-line";
	import RiArrowRightUpLine from "remixicon-svelte/icons/arrow-right-up-line";
	import RiCircleFill from "remixicon-svelte/icons/circle-fill";

	type Props = {
		metricValue: number;
		comparison: MetricComparison;
		format: "percentage" | "duration" | "raw";
	};
	const { metricValue, comparison, format }: Props = $props();

	const getComparisonDelta = (value: number, comp: number) => {
		if (comp === 0) {
			if (value === 0) return 1;
			return value;
		}
		return value / comp;
	};

	const delta = $derived(getComparisonDelta(metricValue, comparison.value));
	const margin = $derived(comparison.averageMargin ?? 0.05);
	type DeltaCategory = "avg" | "above" | "below";
	const category = $derived.by<DeltaCategory>(() => {
		if (delta > 1 + margin) return "above";
		if (delta < 1 - margin) return "below";
		return "avg";
	});

	const deltaIcons: Record<DeltaCategory, Component | undefined> = {
		above: RiArrowRightUpLine,
		below: RiArrowRightDownLine,
		avg: undefined,
	};
	const DeltaIcon = $derived(deltaIcons[category]);
	const deltaText = $derived.by(() => {
		if (category === "avg") return "Average";
		if (comparison.value === 0 || format === "raw") return `${metricValue}`;
		return `${Math.round(Math.abs(delta * 100 - 100))}%`;
	});

	const aboveClasses = "text-destructive border-destructive/80/70";
	const belowClasses = "text-primary border-primary/30";
	const averageClasses = "text-muted-foreground/40 border-muted-foreground/20";
	const categoryClasses = $derived.by(() => {
		if (category === "above") return !!comparison.positive ? belowClasses : aboveClasses;
		if (category === "below") return !!comparison.positive ? aboveClasses : belowClasses;
		return averageClasses;
	});
</script>

<div class="flex flex-col items-center gap-2 py-1 px-2 border rounded-full {categoryClasses}">
	<div class="flex gap-1 text-sm items-center">
		{#if DeltaIcon}<DeltaIcon class="size-[18px]" aria-hidden="true" />{/if}
		<span>{deltaText}</span>
	</div>

	{#if comparison.hint}
		<div class="text-primary">
			<RiCircleFill class="size-4 border rounded-full border-primary" aria-hidden="true" />
			<!--div class="text-sm text-gray-500 mt-1">
				Potential sleep disruptions
			</div-->
		</div>
	{/if}
</div>
