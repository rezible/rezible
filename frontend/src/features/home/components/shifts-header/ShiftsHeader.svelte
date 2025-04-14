<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiPlus } from "@mdi/js";
	import { Button } from "svelte-ux";
	import { getUserOncallInformationOptions } from "$lib/api";
	import { session } from "$lib/auth.svelte";
	import ActiveShiftCard from "./ActiveShiftCard.svelte";
	import WatchRosterDialog from "./WatchRosterDialog.svelte";

	const userId = $derived(session.userId);
	const oncallInfoQuery = createQuery(() => ({
		...getUserOncallInformationOptions({
			query: { userId, activeShifts: true },
		}),
	}));
	const oncallInfo = $derived(oncallInfoQuery.data?.data);
	const rosterIds = $derived(oncallInfo?.rosters.map(r => r.id) ?? []);
	const shifts = $derived(oncallInfoQuery.data?.data.activeShifts);
	const userActiveShift = $derived(shifts?.find(s => (s.attributes.user.id === userId)));

	let rosterDialogOpen = $state(false);
	const onWatchedRostersUpdated = () => {oncallInfoQuery.refetch()};
</script>

<div class="w-full flex gap-2">
	<div class="flex flex-row gap-2 flex-wrap">
		{#if userActiveShift}
			<ActiveShiftCard shift={userActiveShift} />
		{/if}
	</div>
	<div class="grid place-items-center">
		<Button icon={mdiPlus} rounded classes={{root: "h-20 opacity-70 hover:opacity-100"}} on:click={() => (rosterDialogOpen = true)}>
			<span>Pin Roster</span>
		</Button>
	</div>
</div>

<WatchRosterDialog bind:open={rosterDialogOpen} current={rosterIds} onUpdated={onWatchedRostersUpdated} />