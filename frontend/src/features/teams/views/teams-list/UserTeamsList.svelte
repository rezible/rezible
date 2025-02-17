<script lang="ts">
	import { listTeamsOptions, type ListTeamsData, type Team } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import TeamCard from "./TeamCard.svelte";
	import { createQuery } from "@tanstack/svelte-query";

	let params = $state<ListTeamsData>({ query: { limit: 1 } });
	const userTeamsQuery = createQuery(() => listTeamsOptions(params));
</script>

<div class="flex flex-col h-full">
	<LoadingQueryWrapper query={userTeamsQuery}>
		{#snippet view(userTeams: Team[])}
			{#each userTeams as team}
				<TeamCard title={team.attributes.name} teamId={team.id} />
			{/each}
		{/snippet}
	</LoadingQueryWrapper>
</div>
