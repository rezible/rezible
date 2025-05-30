<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import {
	getTeamOptions,
		listOncallRostersOptions,
		listUsersOptions,
		type OncallRoster,
		type User,
	} from "$lib/api";

	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

	import TeamUsers from "./TeamUsers.svelte";
	import TeamRosters from "./TeamRosters.svelte";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import Card from "$components/card/Card.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {
		teamSlug: string;
	}
	let { teamSlug }: Props = $props();

	const teamQuery = createQuery(() => getTeamOptions({ path: { id: teamSlug } }));
	const team = $derived(teamQuery.data?.data);
	const teamId = $derived(team?.id ?? "");
	const teamName = $derived(team?.attributes.name ?? "");
	
	appShell.setPageBreadcrumbs(() => [
		{ label: "Teams", href: "/teams" },
		{ label: teamName, href: `/teams/${teamSlug}`, avatar: { kind: "team", id: teamId } },
	]);

	const usersQuery = createQuery(() => ({...listUsersOptions({ query: { teamId } }), enabled: !!teamId }));
	const rostersQuery = createQuery(() => ({...listOncallRostersOptions({ query: { teamId } }), enabled: !!teamId }));
</script>

<div class="flex gap-2">
	<Card classes={{ root: "max-w-lg", headerContainer: "p-3" }}>
		{#snippet header()}
			<Header title="Users" classes={{title: "text-xl"}} />
		{/snippet}
		{#snippet contents()}
			<LoadingQueryWrapper query={usersQuery}>
				{#snippet view(users: User[])}
					<TeamUsers {users} />
				{/snippet}
			</LoadingQueryWrapper>
		{/snippet}
	</Card>

	<Card classes={{ root: "max-w-lg", headerContainer: "p-3" }}>
		{#snippet header()}
			<Header title="Oncall Rosters" classes={{title: "text-xl"}} />
		{/snippet}
		{#snippet contents()}
			<LoadingQueryWrapper query={rostersQuery}>
				{#snippet view(rosters: OncallRoster[])}
					<TeamRosters {rosters} />
				{/snippet}
			</LoadingQueryWrapper>
		{/snippet}
	</Card>
</div>
