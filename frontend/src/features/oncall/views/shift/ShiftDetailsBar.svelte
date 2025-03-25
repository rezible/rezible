<script lang="ts">
	import { DateToken, PeriodType } from "@layerstack/utils";
	import { isFuture } from "date-fns";
	import { settings } from "$lib/settings.svelte";
	import { shiftState } from "$features/oncall/views/shift/shift.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";

	const attr = $derived(shiftState.shift?.attributes);
	const role = $derived(attr?.role);
	const roster = $derived(attr?.roster);
	const user = $derived(attr?.user);

	const burdenScore = $derived(0.23);
	const burdenRating = $derived("High");

	const startDate = $derived(shiftState.shiftStart?.toDate());
	const endDate = $derived(shiftState.shiftEnd?.toDate());

	const timeFmt = `${DateToken.Hour_numeric}:${DateToken.Minute_numeric}`;
</script>

<div class="flex gap-4 h-14 max-h-14 overflow-y-hidden justify-between">
	<div class="grid grid-flow-col gap-2">
		{#if user && roster}
			<a href="/users/{user.id}" class="flex items-center gap-4 bg-surface-100 rounded-lg hover:bg-accent-800/40 p-2 px-4">
				<Avatar kind="user" size={32} id={user.id} />
				<div class="flex flex-col">
					<span class="text-lg">{user.attributes.name}</span>
				</div>
			</a>

			<a href="/oncall/rosters/{roster.id}" class="flex items-center gap-4 bg-surface-100 rounded-lg hover:bg-accent-800/40 p-2 px-4">
				<Avatar kind="roster" size={32} id={roster.id} />
				<span class="text-lg">{roster.attributes.name}</span>
			</a>
		{/if}
	</div>

	<div class="flex gap-6 border rounded-lg p-2 px-4 w-fit">
		{#snippet formattedDateTime(label: string, d: Date)}
			<div class="flex flex-col text-center">
				<span class="text-sm text-surface-content/60">{label}</span>
				<span class="">
					{settings.format(d, PeriodType.Custom, {custom: timeFmt})}
					<span>{settings.format(d, PeriodType.Custom, {custom: DateToken.DayOfWeek_long})}</span>
					<span>{settings.format(d, PeriodType.Day)}</span>
				</span>
			</div>
		{/snippet}

		{#if startDate && endDate}
			{@render formattedDateTime(isFuture(startDate) ? "Starts" : "Started", startDate)}
			<span class="h-full border-l"></span>
			{@render formattedDateTime(isFuture(endDate) ? "Ends" : "Ended", endDate)}
		{/if}
	</div>
</div>