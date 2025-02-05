<script lang="ts">
	import { getUserOncallDetailsOptions, type UserOncallDetails } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { Header } from "svelte-ux";
	import UserShiftsDisplay from "./UserShiftsDisplay.svelte";
	import UserRostersList from "./UserRostersList.svelte";

	const userOncallQuery = createQuery(() => getUserOncallDetailsOptions());
</script>

<div class="grid grid-cols-4 gap-2 h-full max-h-full">
	<div class="flex flex-col min-h-0 gap-2 max-h-full overflow-hidden">
		<Header title="Your Rosters" subheading="" classes={{ title: "text-2xl" }} />

		<LoadingQueryWrapper query={userOncallQuery}>
			{#snippet view(details: UserOncallDetails)}
				<UserRostersList rosters={details.rosters} />
			{/snippet}
		</LoadingQueryWrapper>
	</div>

	<div class="col-span-3 flex flex-col min-h-0 gap-2 max-h-full">
		<Header title="Your Shifts" subheading="" classes={{ title: "text-2xl" }} />

		<LoadingQueryWrapper query={userOncallQuery}>
			{#snippet view(details: UserOncallDetails)}
				<UserShiftsDisplay
					activeShifts={details.active_shifts}
					upcomingShifts={details.upcoming_shifts}
					pastShifts={details.past_shifts}
				/>
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
