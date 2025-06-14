<script lang="ts" module>
	type MetricValue = number | string;
	export type MetricComparison = {
		value: MetricValue;
		positive?: boolean;
		icon?: string;
		deltaLabel?: string;
		hint?: string;
	};
</script>

<script lang="ts">
	import { mdiArrowBottomRight, mdiArrowBottomRightThin, mdiArrowDownRight, mdiArrowTopRight, mdiArrowTopRightThin, mdiArrowUpRight, mdiArrowUUpRight, mdiCircleMedium, mdiDecimalIncrease } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		metricValue: MetricValue;
		comparison: MetricComparison;
	}
	const { metricValue, comparison }: Props = $props();

	const comparisonClass = (delta: number, positive: boolean) => {
		const low = (positive ? 1 : delta);
		const high = (positive ? delta : 1);
		if (low > high) return "text-danger-500";
		if (low < high) return "text-success-500";
		return "text-neutral-content/40";
	};

	const getComparisonDelta = (value: MetricValue, comp: MetricValue) => {
		if (typeof comp === "string" || typeof value === "string") return 1;
		if (comp === 0) return 1;
		return value / comp;
	};

	const delta = $derived(getComparisonDelta(metricValue, comparison.value));
	const defaultLabel = $derived(delta > 1 ? "above average" : "below average");
	const deltaLabel = $derived(comparison.deltaLabel || defaultLabel);
	const deltaIcon = $derived.by(() => {
		if (delta === 1) return;
		if (delta > 1) return mdiArrowTopRightThin;
		return mdiArrowBottomRightThin;
	})
	const deltaText = $derived.by(() => {
		if (delta === 1) return "Average";
		return `${Math.round(Math.abs((delta * 100) - 100))}%`
	});
</script>

<div class="{comparisonClass(delta, !!comparison.positive)} flex flex-col items-end gap-2">
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