<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import { mdiCircleMedium } from "@mdi/js";
	import { formatDistance } from "date-fns";
	import { Tooltip, ProgressCircle } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		shift: OncallShift;
		pulse?: boolean;
		size?: number;
	};
	const { shift, pulse = true, size = 32 }: Props = $props();

	const start = $derived(new Date(shift.attributes.startAt));
	const end = $derived(new Date(shift.attributes.endAt));
	const progress = $derived(100 * (Date.now() - start.valueOf()) / (end.valueOf() - start.valueOf()));
	const timeLeft = $derived(formatDistance(end, Date.now()));
</script>

<Tooltip>
	<ProgressCircle
		{size}
		value={progress}
		track
		class="text-success [--track-color:theme(colors.success/10%)]"
	>
		{#if pulse}
			<Icon data={mdiCircleMedium} classes={{root: "animate-pulse"}} />
		{/if}
	</ProgressCircle>
	<div
		slot="title"
		class="bg-neutral border text-sm text-surface-content p-2 rounded-lg"
	>
		<div class="flex flex-col gap-2">
			<span>{timeLeft} left</span>
		</div>
	</div>
</Tooltip>