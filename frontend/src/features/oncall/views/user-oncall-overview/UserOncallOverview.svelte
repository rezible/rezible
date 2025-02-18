<script lang="ts">
	import { getUserOncallDetailsOptions, type UserOncallDetails } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { Button, Header, Icon } from "svelte-ux";
	import UserShiftsDisplay from "./UserShiftsDisplay.svelte";
	import UserRostersList from "./UserRostersList.svelte";
	import SplitPage from "$components/split-page/SplitPage.svelte";
	import { mdiArrowRight } from "@mdi/js";

	const userOncallQuery = createQuery(() => getUserOncallDetailsOptions());
</script>

{#snippet rostersNav()}
	<Header title="Rosters" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		<svelte:fragment slot="actions">
			<Button href="/oncall/rosters">
				View All
			</Button>
		</svelte:fragment>
	</Header>

	<LoadingQueryWrapper query={userOncallQuery}>
		{#snippet view(details: UserOncallDetails)}
			<UserRostersList rosters={details.rosters} />
		{/snippet}
	</LoadingQueryWrapper>
{/snippet}

<SplitPage nav={rostersNav}>
	<Header title="Your Shifts" subheading="" classes={{ title: "text-2xl", root: "h-11" }} />

	<LoadingQueryWrapper query={userOncallQuery}>
		{#snippet view(details: UserOncallDetails)}
			<UserShiftsDisplay
				activeShifts={details.activeShifts}
				upcomingShifts={details.upcomingShifts}
				pastShifts={details.pastShifts}
			/>
		{/snippet}
	</LoadingQueryWrapper>
</SplitPage>
