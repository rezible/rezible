<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import type { ZonedDateTime } from "@internationalized/date";
	import { settings } from "$lib/settings.svelte";
	import { DateToken, PeriodType } from "@layerstack/utils";
	import { isFuture } from "date-fns";

	type Props = {
		shift: OncallShift;
		shiftStart: ZonedDateTime;
		shiftEnd: ZonedDateTime;
	};
	let { shift, shiftStart, shiftEnd }: Props = $props();

	const role = $derived(shift.attributes.role);
	const roster = $derived(shift.attributes.roster);
	const user = $derived(shift.attributes.user);

	const burdenScore = $derived(0.23);
	const burdenRating = $derived("High");

	const startDate = $derived(shiftStart.toDate());
	const endDate = $derived(shiftEnd.toDate());

	const timeFmt = `${DateToken.Hour_numeric}:${DateToken.Minute_numeric}`;
</script>

<div class="flex gap-4 h-14 max-h-14 overflow-y-hidden justify-between">
	<div class="grid grid-flow-col gap-2">
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

		{@render formattedDateTime(isFuture(startDate) ? "Starts" : "Started", startDate)}
		<span class="h-full border-l"></span>
		{@render formattedDateTime(isFuture(endDate) ? "Ends" : "Ended", endDate)}
	</div>
</div>