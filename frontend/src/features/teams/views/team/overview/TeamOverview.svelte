<script lang="ts">
	import { listOncallRostersOptions, listUsersOptions, type OncallRoster, type User } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import Card from "$components/card/Card.svelte";
	import Header from "$components/header/Header.svelte";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { useTeamViewState } from "../viewState.svelte";
	import { createQuery } from "@tanstack/svelte-query";

	const viewState = useTeamViewState();

	const teamId = $derived(viewState.teamId);
	const usersQuery = createQuery(() => ({...listUsersOptions({ query: { teamId } }) }));
	const rostersQuery = createQuery(() => ({...listOncallRostersOptions({ query: { teamId } }) }));
</script>

<div class="flex gap-2">
	<Card classes={{ root: "max-w-lg pb-3", headerContainer: "p-3" }}>
		{#snippet header()}
			<Header title="Users" classes={{title: "text-xl"}} />
		{/snippet}
		{#snippet contents()}
			<LoadingQueryWrapper query={usersQuery}>
				{#snippet view(users: User[])}
					<div class="flex flex-col gap-2 w-fit">
						{#each users as user (user.id)}
							<a href="/users/{user.id}" class="">
								<div class="flex items-center gap-2 px-2 hover:bg-surface-200 p-1 rounded-lg w-full">
									<Avatar kind="user" id={user.id} size={32} />
									<span>{user.attributes.name}</span>
								</div>
							</a>
						{/each}
					</div>
				{/snippet}
			</LoadingQueryWrapper>
		{/snippet}
	</Card>

	<Card classes={{ root: "max-w-lg pb-3", headerContainer: "p-3" }}>
		{#snippet header()}
			<Header title="Oncall Rosters" classes={{title: "text-xl"}} />
		{/snippet}
		{#snippet contents()}
			<LoadingQueryWrapper query={rostersQuery}>
				{#snippet view(rosters: OncallRoster[])}
					<div class="flex flex-col gap-2 w-fit">
						{#each rosters as roster (roster.id)}
							<a href="/oncall/rosters/{roster.id}">
								<div class="flex items-center gap-2 px-2 hover:bg-surface-200 p-1 rounded-lg">
									<Avatar id={roster.id} size={32} kind="roster" />
									<span>{roster.attributes.name}</span>
								</div>
							</a>
						{/each}
					</div>
				{/snippet}
			</LoadingQueryWrapper>
		{/snippet}
	</Card>
</div>
