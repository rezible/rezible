<script lang="ts">
	import { getUserOncallInformationOptions, type UserOncallInformation } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { Button } from "svelte-ux";
	import UserShiftsDisplay from "./UserShiftsDisplay.svelte";
	import UserRostersList from "./UserRostersList.svelte";
	import SplitPage from "$components/split-page/SplitPage.svelte";
	import Header from "$src/components/header/Header.svelte";

	const userOncallQuery = createQuery(() => getUserOncallInformationOptions());
</script>

<SplitPage>
	{#snippet nav()}
		<Header title="Rosters" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
			{#snippet actions()}
				<Button href="/oncall/rosters">View All</Button>
			{/snippet}
		</Header>

		<LoadingQueryWrapper query={userOncallQuery}>
			{#snippet view(details: UserOncallInformation)}
				<UserRostersList rosters={details.rosters} />
			{/snippet}
		</LoadingQueryWrapper>
	{/snippet}

	<Header title="Shifts" subheading="" classes={{ title: "text-2xl", root: "h-11" }}>
		{#snippet actions()}
			<Button href="/oncall/shifts">
				<span>View All</span>
			</Button>
		{/snippet}
	</Header>

	<LoadingQueryWrapper query={userOncallQuery}>
		{#snippet view(info: UserOncallInformation)}
			<UserShiftsDisplay {info} />
		{/snippet}
	</LoadingQueryWrapper>
</SplitPage>
