<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getUserOncallInformationOptions } from "$lib/api";
	import { session } from "$lib/auth.svelte";

	import ShiftsHeader from "$features/home/components/shifts-header/ShiftsHeader.svelte";
	import UserItems from "$features/home/components/user-items/UserItems.svelte";
	import UserPinnedItems from "$features/home/components/user-items/UserPinnedItems.svelte";
	import OncallEventsTable from "$components/oncall-events-table/OncallEventsTable.svelte";

	const userId = $derived(session.userId);
	const oncallInfoQuery = createQuery(() => getUserOncallInformationOptions({query: { userId }}));
	const oncallInfo = $derived(oncallInfoQuery.data?.data);
	const userActiveShifts = $derived(oncallInfo?.activeShifts ?? []);
	const userRosterIds = $derived(oncallInfo?.rosters.map(r => r.id) ?? []);

	const getDefaultRosterIds = () => {
		if (userActiveShifts.length === 0) {
			if (userRosterIds.length === 0) return undefined;
			return userRosterIds;
		};
		return userActiveShifts.map(s => s.attributes.roster.id);
	};
	const defaultRosterIds = $derived.by(getDefaultRosterIds);
</script>

<div class="h-full w-full flex flex-col gap-2">
	<ShiftsHeader {oncallInfo} />

	<div class="flex-1 flex min-h-0 gap-2">
		<div class="flex-1">
			<OncallEventsTable allowRosterActions={userRosterIds} defaultFilters={{rosterIds: defaultRosterIds}} />
		</div>
		<div class="w-2/5 flex flex-col gap-2 min-h-0">
			<UserItems />
			<UserPinnedItems />
		</div>
	</div>
</div>
