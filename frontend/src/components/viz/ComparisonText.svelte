<script lang="ts">
	import { Icon } from "svelte-ux";
	import { formatDelta } from "$lib/format.svelte";
	import { mdiCircleMedium } from "@mdi/js";
	import { comparisonClass, type MetricComparison } from "./utils";

	type Props = {
		comparison: MetricComparison;
	}

	const { comparison }: Props = $props();
</script>

<div class="{comparisonClass(comparison)} flex items-center gap-2">
	{formatDelta(comparison.value)} {comparison.deltaLabel ?? "from average"}

	{#if comparison.hint}
		<div class="text-warning">
			<Icon data={mdiCircleMedium} size={16} classes={{root: "border rounded-full border-warning"}} />
			<!--div class="text-sm text-gray-500 mt-1">
				Potential sleep disruptions
			</div-->
		</div>
	{/if}
</div>