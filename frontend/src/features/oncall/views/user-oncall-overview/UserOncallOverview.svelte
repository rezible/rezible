<script lang="ts">
	import { getUserOncallInformationOptions, type UserOncallInformation } from "$lib/api";
	import { createQuery } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { Button } from "svelte-ux";
	import UserShiftsDisplay from "./UserShiftsDisplay.svelte";
	import UserRostersList from "./UserRostersList.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SectionHeader from "$components/section-header/SectionHeader.svelte";

	const userOncallQuery = createQuery(() => getUserOncallInformationOptions());
</script>

{#snippet filters()}
	<SectionHeader title="Your Rosters">
		{#snippet actions()}
			<Button href="/rosters">View All</Button>
		{/snippet}
	</SectionHeader>

	<LoadingQueryWrapper query={userOncallQuery}>
		{#snippet view(details: UserOncallInformation)}
			<UserRostersList rosters={details.rosters} />
		{/snippet}
	</LoadingQueryWrapper>
{/snippet}

<FilterPage {filters}>
	<SectionHeader title="Your Shifts">
		{#snippet actions()}
			<Button href="/shifts">View All</Button>
		{/snippet}
	</SectionHeader>

	<LoadingQueryWrapper query={userOncallQuery}>
		{#snippet view(info: UserOncallInformation)}
			<UserShiftsDisplay {info} />
		{/snippet}
	</LoadingQueryWrapper>
</FilterPage>
