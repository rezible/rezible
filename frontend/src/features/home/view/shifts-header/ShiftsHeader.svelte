<script lang="ts">
	import { mdiPlus } from "@mdi/js";
	import Button from "$components/button/Button.svelte";
	import { type OncallShift } from "$lib/api";
	import { useAuthSessionState } from "$lib/auth.svelte";
	import ActiveShiftCard from "./ActiveShiftCard.svelte";
	import WatchRosterDialog from "./WatchRosterDialog.svelte";
	import { useUserOncallInformation } from "$lib/userOncall.svelte";

	const session = useAuthSessionState();
	const userId = $derived(session.user?.id);
	const oncallInfo = useUserOncallInformation();

	const watchedRosterIds = $derived(oncallInfo.current?.watchingRosters.map(r => r.id) ?? []);
	const userRosterIds = $derived(oncallInfo.rosterIds);
	const rosterIds = $derived([...userRosterIds, ...watchedRosterIds]);

	const shifts = $derived(oncallInfo.activeShifts);
	const [userShifts, rosterShifts] = $derived.by(() => {
		let userShifts: OncallShift[] = [];
		let rosterShifts: OncallShift[] = [];

		// const watchedRosterIdsSet = $derived(new Set(watchedRosterIds));
		shifts.forEach(s => {
			if (s.attributes.user.id === userId) {
				userShifts.push(s);
			} else {
				rosterShifts.push(s);
			}
		});

		return [userShifts, rosterShifts];
	});

	let rosterDialogOpen = $state(false);
	const onWatchedRostersUpdated = () => oncallInfo.invalidate();
</script>

<div class="w-full flex gap-2">
	<div class="flex flex-row gap-2 flex-wrap">
		{#each userShifts as shift, i}
			<ActiveShiftCard {shift} isUser />
		{/each}
		{#each rosterShifts as shift, i}
			<ActiveShiftCard {shift}  />
		{/each}
	</div>
	<div class="grid place-items-center">
		<Button icon={mdiPlus} rounded classes={{root: "h-20 opacity-70 hover:opacity-100"}} onclick={() => (rosterDialogOpen = true)}>
			<span>Watch Roster</span>
		</Button>
	</div>
</div>

<WatchRosterDialog bind:open={rosterDialogOpen} current={rosterIds} onUpdated={onWatchedRostersUpdated} />