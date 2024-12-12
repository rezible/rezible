<script lang="ts">
    import { createQuery } from "@tanstack/svelte-query";
    import { listOncallShiftsOptions } from "$lib/api";
    import { session } from "$lib/auth.svelte";
    import ShiftCard from "$features/home/components/shift-card/ShiftCard.svelte";
    import UserItems from "$features/home/components/user-items/UserItems.svelte";
    import UserPinnedItems from "$features/home/components/user-items/UserPinnedItems.svelte";
    import TeamInfo from "$features/home/components/events-overview/TeamInfo.svelte";

	const userShiftsQuery = createQuery(() => ({...listOncallShiftsOptions({query: {user_id: session.userId, active: true}}), enabled: !!session.userId}));
	const currentShifts = $derived(userShiftsQuery.data?.data);
</script>

<div class="h-full w-full grid grid-cols-2 gap-3 min-h-0 max-h-full gap-2">
	<div class="grid auto-rows-min gap-2 min-h-0">
		<UserItems />
		<UserPinnedItems />
	</div>

	{#if currentShifts}
		<div class="max-h-full min-h-0 inline">
			<ShiftCard shifts={currentShifts} />

			<div class="mt-2">
				<TeamInfo />
			</div>
		</div>
	{/if}
</div>