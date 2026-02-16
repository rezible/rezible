<script lang="ts">
import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
	import { listOncallShiftsOptions } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import Header from "$components/header/Header.svelte";
	import { mdiArrowRight } from "@mdi/js";
	import ShiftCard from "$features/oncall/components/shift-card/ShiftCard.svelte";
	import { useOncallRosterViewController } from "$features/oncall/views/roster";
	
	const view = useOncallRosterViewController();
	const rosterId = $derived(view.rosterId);

	// TODO: use correct query
	const shiftsQuery = createQuery(() => listOncallShiftsOptions({ query: {userId: rosterId} }));
	const shifts = $derived(shiftsQuery.data?.data);
	const prevShift = $derived(shifts?.at(0));
	const activeShift = $derived(shifts?.at(0));
	const nextShift = $derived(shifts?.at(0));
</script>

<div class="flex flex-col h-full border border-surface-content/10 rounded">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Shifts" classes={{root: "", title: "text-xl"}}>
			{#snippet actions()}
				<Button href={`/rosters/${rosterId}/shifts`}>
					View All
					<Icon data={mdiArrowRight} classes={{root: "ml-1 h-4 w-4"}} />
				</Button>
			{/snippet}
		</Header>
	</div>

	<div class="flex-1 flex flex-col gap-2 px-0 overflow-y-auto">
		{#if nextShift}<ShiftCard shift={nextShift} hideRoster />{/if}
		{#if activeShift}<ShiftCard shift={activeShift} hideRoster />{/if}
		{#if prevShift}<ShiftCard shift={prevShift} hideRoster />{/if}
	</div>
</div>