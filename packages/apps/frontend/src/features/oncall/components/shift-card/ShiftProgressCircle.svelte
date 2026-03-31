<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import { formatDistanceToNow } from "date-fns";

	type Props = {
		shift: OncallShift;
		pulse?: boolean;
		size?: number;
	};
	const { shift, pulse = true, size = 32 }: Props = $props();

	const start = $derived(new Date(shift.attributes.startAt));
	const end = $derived(new Date(shift.attributes.endAt));
	const progress = $derived(100 * (Date.now() - start.valueOf()) / (end.valueOf() - start.valueOf()));
	const timeLeft = $derived(formatDistanceToNow(end));
</script>

<!--Tooltip>
	<ProgressCircle
		{size}
		width={size/4}
		value={progress}
		track
		class={cls("text-success [--track-color:theme(colors.success/10%)]")}
	>
	</ProgressCircle>
	<div
		slot="title"
		class="bg-neutral border text-sm text-surface-content p-2 rounded-lg"
	>
		<div class="flex flex-col gap-2">
			<span>{timeLeft} left</span>
		</div>
	</div>
</Tooltip-->