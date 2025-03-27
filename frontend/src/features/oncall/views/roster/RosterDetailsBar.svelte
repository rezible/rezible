<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import { getUserOncallDetailsOptions } from "$src/lib/api";
	import ShiftProgressCircle from "$features/oncall/components/shift-progress-circle/ShiftProgressCircle.svelte";
	import type { OncallRoster } from "./types";
	import { Icon } from "svelte-ux";
	import { mdiSlack } from "@mdi/js";

	const shiftsQuery = createQuery(() => getUserOncallDetailsOptions());
	const shifts = $derived(shiftsQuery.data?.data);

	const activeShift = $derived(shifts?.activeShifts.at(0));

	const userLocalTime = new Date(Date.now()).toLocaleTimeString(undefined, {hour: "2-digit", minute: "2-digit", hour12: true})
</script>

<div class="flex gap-4 h-14 max-h-14 overflow-y-hidden justify-between pb-2">
	<!-- <a href="/oncall/" class="flex items-center gap-4 px-4 bg-accent-900/50 rounded-lg hover:bg-accent-900/40">
		<div class="flex flex-col">
			<span class="text-xs">Slack Channel</span>
			<div class="flex items-center gap-2">
				<Icon data={mdiSlack} size={16} />
				<span class="">#{slackChannel}</span>
			</div>
		</div>
	</a> -->
	
	{#if activeShift}
		<a href="/oncall/shifts/{activeShift.id}" class="flex items-center gap-4 px-4 bg-success-900/50 rounded-lg hover:bg-success-900/40">
			<div class="flex flex-col">
				<span class="text-xs">Currently Oncall</span>
				<div class="flex items-center align-middle gap-2">
					<Avatar kind="user" size={14} id={activeShift.id} />
					<span class="text-sm font-semibold">{activeShift.attributes.user.attributes.name}</span>
					<span class="text-xs text-surface-content/70 align-middle">({userLocalTime})</span>
				</div>
			</div>
			<div class="">
				<ShiftProgressCircle shift={activeShift} size={30} pulse={false} />
			</div>
		</a>
	{/if}
</div>