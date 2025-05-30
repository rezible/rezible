<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { TextField, ListItem, Button } from "svelte-ux";
	import { mdiMagnify, mdiChevronRight } from "@mdi/js";
	import { listOncallRostersOptions, type ListOncallRostersData, type OncallRoster } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import SplitPage from "$src/components/split-page/SplitPage.svelte";
	import UserRosterCard from "./UserRosterCard.svelte";
	import Header from "$src/components/header/Header.svelte";

	let allParams = $state<ListOncallRostersData>();
	const allQuery = createQuery(() => listOncallRostersOptions(allParams));
	const allRosters = $derived(allQuery.data?.data);

	let userParams = $state<ListOncallRostersData>();
	const userQuery = createQuery(() => listOncallRostersOptions(userParams));
	const userRosters = $derived(userQuery.data?.data);

	const updateSearch = (value: any) => {
		console.log(value);
	};
</script>

<SplitPage>
	{#snippet nav()}
		<Header title="Your Rosters" subheading="" classes={{ title: "text-2xl", root: "h-11" }} />

		<div class="flex flex-col h-full">
			<LoadingQueryWrapper query={userQuery}>
				{#snippet view(rosters: OncallRoster[])}
					{#each rosters as r (r.id)}
						<UserRosterCard
							title={r.attributes.name}
							href="/oncall/rosters/{r.attributes.slug}"
							rosterId={r.id}
						/>
					{/each}
				{/snippet}
			</LoadingQueryWrapper>
		</div>
	{/snippet}

	<div class="flex flex-col h-full gap-2 overflow-x-hidden overflow-y-auto">
		<div class="">
			<TextField
				dense
				rounded
				label="Search For Rosters"
				labelPlacement="float"
				icon={mdiMagnify}
				debounceChange={500}
				on:change={({ detail }) => console.log(detail)}
			/>
		</div>
	
		<LoadingQueryWrapper query={allQuery}>
			{#snippet view(rosters: OncallRoster[])}
				<div class="min-h-0 flex flex-col gap-2 overflow-y-auto flex-1 px-2">
					{#each rosters as r}
						<a href="/oncall/rosters/{r.attributes.slug}">
							<ListItem title={r.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
								<svelte:fragment slot="avatar">
									<Avatar kind="roster" size={32} id={r.id} />
								</svelte:fragment>
								<div slot="actions">
									<Button icon={mdiChevronRight} class="p-2 text-surface-content/50" />
								</div>
							</ListItem>
						</a>
					{/each}
				</div>
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</SplitPage>

