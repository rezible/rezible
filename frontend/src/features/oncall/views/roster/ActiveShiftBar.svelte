<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import { getUserOncallDetailsOptions } from "$src/lib/api";
	import ShiftProgressCircle from "$features/oncall/components/shift-progress-circle/ShiftProgressCircle.svelte";

	const shiftsQuery = createQuery(() => getUserOncallDetailsOptions());
	const shifts = $derived(shiftsQuery.data?.data);

	const activeShift = $derived(shifts?.activeShifts.at(0));
</script>

<div class="flex gap-4 h-14 max-h-14 overflow-y-hidden justify-between pb-2">
	{#if activeShift}
		<a href="/oncall/shifts/{activeShift.id}" class="flex items-center gap-4 px-4 bg-success-900/50 rounded-lg hover:bg-success-900/40">
			<div class="flex flex-col">
				<span class="text-xs">Currently Oncall</span>
				<div class="flex items-center gap-2">
					<Avatar kind="user" size={14} id={activeShift.id} />
					<span class="text-sm font-semibold">{activeShift.attributes.user.attributes.name}</span>
				</div>
			</div>
			<div class="">
				<ShiftProgressCircle shift={activeShift} size={30} pulse={false} />
			</div>
		</a>
	{/if}
</div>