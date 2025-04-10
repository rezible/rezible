<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { listOncallShiftsOptions } from "$lib/api";
	import { session } from "$lib/auth.svelte";
	import ActiveShiftCard from "./ActiveShiftCard.svelte";
	import { mdiPlus } from "@mdi/js";
	import { Button } from "svelte-ux";

	const userId = $derived(session.userId);
	const pinnedRosters = $state<string[]>([]);
	const shiftsQuery = createQuery(() => ({
		...listOncallShiftsOptions({
			query: { userId, active: true },
		}),
	}));
	const shifts = $derived(shiftsQuery.data?.data);
	const userActiveShift = $derived(shifts?.find(s => (s.attributes.user.id === userId)));
</script>

<div class="w-full flex gap-2">
	<div class="flex flex-row gap-2 flex-wrap">
		{#if userActiveShift}
			<ActiveShiftCard shift={userActiveShift} />
		{/if}
	</div>
	<div class="grid place-items-center">
		<Button icon={mdiPlus} rounded={false} classes={{root: "h-20 opacity-70 hover:opacity-100"}}>
			<span>Pin Roster</span>
		</Button>
	</div>
</div>