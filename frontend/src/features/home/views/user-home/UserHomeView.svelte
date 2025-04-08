<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listOncallShiftsOptions } from "$lib/api";
	import { session } from "$lib/auth.svelte";
	import ShiftCard from "$features/home/components/shift-card/ShiftCard.svelte";
	import UserItems from "$features/home/components/user-items/UserItems.svelte";
	import UserPinnedItems from "$features/home/components/user-items/UserPinnedItems.svelte";
	import EventsOverview from "$features/home/components/events-overview/EventsOverview.svelte";

	const userShiftsQuery = createQuery(() => ({
		...listOncallShiftsOptions({
			query: { userId: session.userId, active: true },
		}),
		enabled: !!session.userId,
	}));
	const activeShift = $derived(userShiftsQuery.data?.data.at(0));
	// const currentlyOncall = $derived(!!activeShift)
</script>

<div class="h-full w-full flex flex-col gap-2">
	{#if activeShift}
		<div class="w-full">
			<ShiftCard shift={activeShift} />
		</div>
	{/if}

	<div class="flex-1 flex min-h-0 gap-2">
		<div class="flex-1">
			<EventsOverview {activeShift} />
		</div>
		<div class="w-1/3 flex flex-col gap-2 min-h-0">
			<UserItems />
			<UserPinnedItems />
		</div>
	</div>
</div>
