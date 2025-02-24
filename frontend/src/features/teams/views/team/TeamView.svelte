<script lang="ts">
	import { Card } from "svelte-ux";
	import { createQuery } from "@tanstack/svelte-query";
	import {
	getTeamOptions,
		listOncallRostersOptions,
		listUsersOptions,
		type OncallRoster,
		type Team,
		type User,
	} from "$lib/api";

	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

	import TeamUsers from "./TeamUsers.svelte";
	import TeamRosters from "./TeamRosters.svelte";

	type Props = {
		teamId: string;
	}
	let { teamId }: Props = $props();

	const teamQuery = createQuery(() => getTeamOptions({ path: { id: teamId } }));
	const usersQuery = createQuery(() => listUsersOptions({ query: { teamId } }));
	const rostersQuery = createQuery(() => listOncallRostersOptions({ query: { teamId } }));
</script>

<div class="flex gap-2">
	<Card title="Users" class="max-w-lg" classes={{ header: { title: "text-xl" }, headerContainer: "p-3" }}>
		<div slot="contents">
			<LoadingQueryWrapper query={usersQuery}>
				{#snippet view(users: User[])}
					<TeamUsers {users} />
				{/snippet}
			</LoadingQueryWrapper>
		</div>
		<div slot="actions"></div>
	</Card>

	<Card
		title="Oncall Rosters"
		class="max-w-lg"
		classes={{ header: { title: "text-xl" }, headerContainer: "p-3" }}
	>
		<div slot="contents">
			<LoadingQueryWrapper query={rostersQuery}>
				{#snippet view(rosters: OncallRoster[])}
					<TeamRosters {rosters} />
				{/snippet}
			</LoadingQueryWrapper>
		</div>
		<div slot="actions"></div>
	</Card>

	<!--Card title="Owned Services" class="max-w-lg" classes={{header: {title: "text-xl"}, headerContainer: "p-3"}}>
		<div slot="contents">
			<LoadingQueryWrapper query={servicesQuery}>
				{#snippet view(services: Service[])}
					<TeamServices {services} />
				{/snippet}
			</LoadingQueryWrapper>
		</div>
		<div slot="actions">
			
		</div>
	</Card-->
</div>
