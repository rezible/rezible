<script lang="ts">
	import { Button } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { formatDistanceToNow } from "date-fns";
	import { rosterViewCtx } from "../viewState.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import {
		getUserOncallInformationOptions,
		type OncallShift,
	} from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import { cls } from "@layerstack/tailwind";
	import { parseAbsoluteToLocal } from "@internationalized/date";
	import Header from "$src/components/header/Header.svelte";
	import { mdiArrowRight } from "@mdi/js";
	import ShiftCard from "$features/oncall-shifts-list/components/shift-card/ShiftCard.svelte";

	const viewCtx = rosterViewCtx.get();
	const rosterId = $derived(viewCtx.rosterId);

	// TODO: use correct query
	const shiftsQuery = createQuery(() => getUserOncallInformationOptions({ query: {} }));
	const shifts = $derived(shiftsQuery.data?.data);
	const prevShift = $derived(shifts?.pastShifts.at(0));
	const activeShift = $derived(shifts?.activeShifts.at(0));
	const nextShift = $derived(shifts?.upcomingShifts.at(0));
</script>

<div class="flex flex-col h-full border border-surface-content/10 rounded">
	<div class="h-fit p-2 flex flex-col gap-2">
		<Header title="Shifts" classes={{root: "", title: "text-xl"}}>
			{#snippet actions()}
				<Button variant="fill-light" href={`/rosters/${rosterId}/shifts`}>
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