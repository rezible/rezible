<script lang="ts">
	import { isFuture } from "date-fns";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { useOncallShiftViewState } from "$features/oncall-shift";

	const view = useOncallShiftViewState();
	const attr = $derived(view.shift?.attributes);
	const roster = $derived(attr?.roster);
	const user = $derived(attr?.user);

	const startDate = $derived(view.shiftStart?.toDate());
	const endDate = $derived(view.shiftEnd?.toDate());

	// const timeFmt = `${DateToken.Hour_numeric}:${DateToken.Minute_numeric}`;
</script>

<div class="grid grid-flow-col gap-2">
	{#if user && roster}
		<a href="/users/{user.id}" class="flex items-center gap-2 bg-surface-100 rounded-lg hover:bg-accent-800/40 p-1 px-3">
			<Avatar kind="user" size={24} id={user.id} />
			<div class="flex flex-col">
				<span class="text-lg">{user.attributes.name}</span>
			</div>
		</a>

		<a href="/rosters/{roster.id}" class="flex items-center gap-2 bg-surface-100 rounded-lg hover:bg-accent-800/40 p-1 px-3">
			<Avatar kind="roster" size={24} id={roster.id} />
			<span class="text-lg">{roster.attributes.name}</span>
		</a>
	{/if}
</div>

<div class="flex gap-2 border rounded-lg p-1 px-4 w-fit items-center">
	{#snippet formattedDateTime(label: string, d: Date)}
		<span class="leading-none">
			<!-- {settings.format(d, PeriodType.Day)}
			{settings.format(d, PeriodType.Custom, {custom: timeFmt})} -->
		</span>
	{/snippet}

	{#if startDate && endDate}
		{@render formattedDateTime(isFuture(startDate) ? "Starts" : "Started", startDate)}
		<span class="leading-none">-</span>
		{@render formattedDateTime(isFuture(endDate) ? "Ends" : "Ended", endDate)}
	{/if}
</div>