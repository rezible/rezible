<script lang="ts" module>
	export type MetricComparison = {
		value: number;
		averageMargin?: number;
		positive?: boolean;
		icon?: string;
		deltaLabel?: string;
		hint?: string;
	};
</script>

<script lang="ts">
	import { mdiArrowBottomRightThin, mdiArrowTopRightThin, mdiCircleMedium } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		metricValue: number;
		comparison: MetricComparison;
		format: "percentage" | "duration" | "raw";
	}
	const { metricValue, comparison, format }: Props = $props();

	const getComparisonDelta = (value: number, comp: number) => {
		if (comp === 0) {
			if (value === 0) return 1;
			return value;
		};
		return value / comp;
	};

	const delta = $derived(getComparisonDelta(metricValue, comparison.value));
	const margin = $derived(comparison.averageMargin ?? .05);
	type DeltaCategory = "avg" | "above" | "below";
	const category = $derived.by<DeltaCategory>(() => {
		if (delta > (1 + margin)) return "above";
		if (delta < (1 - margin)) return "below";
		return "avg";
	});

	const deltaIcons: Record<DeltaCategory, string> = {
		"above": mdiArrowTopRightThin,
		"below": mdiArrowBottomRightThin,
		"avg": "",
	};
	const deltaIcon = $derived(deltaIcons[category]);
	const deltaText = $derived.by(() => {
		if (category === "avg") return "Average";
		if (comparison.value === 0 || format === "raw") return `${metricValue}`;
		return `${Math.round(Math.abs((delta * 100) - 100))}%`;
	});

	const aboveClasses = "text-danger-500 border-danger-900/70";
	const belowClasses = "text-success-500 border-success-900";
	const averageClasses = "text-neutral-content/40 border-neutral-content/20";
	const categoryClasses = $derived.by(() => {
		if (category === "above") return !!comparison.positive ? belowClasses : aboveClasses;
		if (category === "below") return !!comparison.positive ? aboveClasses : belowClasses;
		return averageClasses;
	})
</script>

<div 
	class="flex flex-col items-center gap-2 py-1 px-2 border rounded-full {categoryClasses}"
>
	<div class="flex gap-1 text-sm items-center">
		{#if deltaIcon}<Icon data={deltaIcon} size={18} />{/if}
		<span>{deltaText}</span>
	</div>

	{#if comparison.hint}
		<div class="text-warning">
			<Icon data={mdiCircleMedium} size={16} classes={{root: "border rounded-full border-warning"}} />
			<!--div class="text-sm text-gray-500 mt-1">
				Potential sleep disruptions
			</div-->
		</div>
	{/if}
</div>