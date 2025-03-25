<script lang="ts">
	import { Icon } from "svelte-ux";
	import { formatDelta } from "$lib/format.svelte";
	import { mdiCircleMedium } from "@mdi/js";
	import type { Snippet } from "svelte";
	import InlineStat, { type InlineStatProps } from "./InlineStat.svelte";

	type Props = {
		chart: Snippet;
		stats: InlineStatProps[];
		reverse?: boolean;
	}

	const { chart, stats, reverse }: Props = $props();
</script>

<div class="flex gap-4" class:flex-row-reverse={reverse}>
	<div class="w-1/3 grid place-items-center">
		{@render chart()}
	</div>

	<div class="flex-1 border rounded-lg flex flex-col divide-y">
		{#each stats as props}
			<InlineStat {...props} />
		{/each}
	</div>
</div>