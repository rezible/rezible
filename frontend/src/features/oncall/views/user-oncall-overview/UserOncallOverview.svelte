<script lang="ts">
	import { getUserOncallDetailsOptions, type UserOncallDetails } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { Button, Header, Icon } from "svelte-ux";
	import UserShiftsDisplay from "./UserShiftsDisplay.svelte";
	import UserRostersList from "./UserRostersList.svelte";
	import SplitPage from "$components/split-page/SplitPage.svelte";

	const userOncallQuery = createQuery(() => getUserOncallDetailsOptions());
</script>

<SplitPage>
	{#snippet nav()}
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

	<Header title="Shifts" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		<svelte:fragment slot="actions">
			<Button href="/oncall/shifts">
				<span>View All</span>
			</Button>
		</svelte:fragment>
	</Header>

	<LoadingQueryWrapper query={userOncallQuery}>
		{#snippet view(details: UserOncallDetails)}
			<UserShiftsDisplay {details} />
		{/snippet}
	</LoadingQueryWrapper>
</SplitPage>
