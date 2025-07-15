<script lang="ts">
	import { DateToken, PeriodType } from "@layerstack/utils";
	import { isFuture } from "date-fns";
	import { settings } from "$lib/settings.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { useShiftViewState } from "./shiftViewState.svelte";

	const viewState = useShiftViewState();
	const attr = $derived(viewState.shift?.attributes);
	const roster = $derived(attr?.roster);
	const user = $derived(attr?.user);

	const startDate = $derived(viewState.shiftStart?.toDate());
	const endDate = $derived(viewState.shiftEnd?.toDate());

	const timeFmt = `${DateToken.Hour_numeric}:${DateToken.Minute_numeric}`;
</script>

<div class="flex gap-4 h-12 max-h-14 overflow-y-hidden justify-between pb-2">
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
				{settings.format(d, PeriodType.Day)}
				{settings.format(d, PeriodType.Custom, {custom: timeFmt})}
			</span>
		{/snippet}

		{#if startDate && endDate}
			{@render formattedDateTime(isFuture(startDate) ? "Starts" : "Started", startDate)}
			<span class="leading-none">-</span>
			{@render formattedDateTime(isFuture(endDate) ? "Ends" : "Ended", endDate)}
		{/if}
	</div>
</div>