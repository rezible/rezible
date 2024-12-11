<script lang="ts">
    import { Header, Icon } from "svelte-ux";
    import { mdiGhost } from "@mdi/js";
    import ShiftCard from "$features/home/components/shift-card/ShiftCard.svelte";
    import UserItems from "$features/home/components/user-items/UserItems.svelte";
    import TeamInfo from "$features/home/components/events-overview/TeamInfo.svelte";
    import { listOncallShiftsOptions } from "$src/lib/api";
    import { createQuery } from "@tanstack/svelte-query";
    import { session } from "$src/lib/auth.svelte";

	const userShiftsQuery = createQuery(() => ({...listOncallShiftsOptions({query: {user_id: session.userId, active: true}}), enabled: !!session.userId}));
	const currentShifts = $derived(userShiftsQuery.data?.data);
</script>

<div class="h-full w-full grid grid-cols-2 gap-2 min-h-0 max-h-full gap-2">
	<div class="grid auto-rows-min gap-2 min-h-0">
		<UserItems />

		<div class="flex flex-col gap-2 rounded-lg p-2 border">
			<Header title="Pinned Items" class="border-b" />
			
			<span class="flex items-center gap-2">Nothing Here <Icon data={mdiGhost} /></span>
		</div>
	</div>

	{#if currentShifts}
		<div class="max-h-full min-h-0 inline">
			<ShiftCard shifts={currentShifts} />

			<div class="mt-2">
				<TeamInfo />
			</div>
		</div>
	{/if}
<!-- 
	<div class="">
		<TeamInfo />
	</div> -->
</div>