<script lang="ts">
	import { Header, Icon } from "svelte-ux";
	import {
		mdiChevronRight,
		mdiAlarmLight,
		mdiFire,
		mdiSleepOff,
	} from "@mdi/js";
	import type { OncallShift } from "$src/lib/api";

	type Props = {
		shift: OncallShift;
	};
	const { shift }: Props = $props();
</script>

<div
	class="flex flex-col gap-2 flex-1 min-h-0 max-h-full overflow-y-auto border rounded-lg p-2"
>
	<Header title="Stats" />

	<div class="flex flex-row gap-4">
		<div
			class="flex items-center gap-4 bg-info-900/80 h-full p-2 px-6 rounded-xl w-fit"
		>
			<div class="">
				<Icon data={mdiAlarmLight} />
			</div>
			<div class="flex flex-col">
				<span class="text-lg">25 Alerts</span>
				<span class="text-surface-content/75">Normal</span>
			</div>
		</div>
		<div
			class="flex items-center gap-4 bg-accent-900/80 h-full p-3 rounded-xl w-fit"
		>
			<div class="">
				<Icon data={mdiSleepOff} />
			</div>
			<div class="flex flex-col">
				<span class="text-lg">6 Alerts at Night</span>
				<span class="text-surface-content/75">Above Average</span>
			</div>
		</div>
		<div
			class="flex items-center gap-4 bg-warning-900/80 h-full p-3 rounded-xl w-fit"
		>
			<div class="">
				<Icon data={mdiFire} />
			</div>
			<div class="flex flex-col">
				<span class="text-lg">2 Incidents</span>
				<span class="text-surface-content/75">Above Average</span>
			</div>
		</div>
	</div>

	<Header title="Disruption by Hour" />

	<div class="grid grid-cols-2">
		<div class="grid grid-cols-12">
			{#each new Array(12) as _, hour}
				<div class="border h-8">{hour === 0 ? "12" : hour}am</div>
			{/each}
		</div>
		<div class="grid grid-cols-12">
			{#each new Array(12) as _, hour}
				<div class="border h-8">{hour === 0 ? "12" : hour}pm</div>
			{/each}
		</div>
	</div>
</div>
