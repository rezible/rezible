<script lang="ts">
	import { listTeamsOptions, type ListTeamsData, type Team } from '$lib/api';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';
	import TeamCard from './TeamCard.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	
	let params = $state<ListTeamsData>({query: {limit: 1}});
	const userTeamsQuery = createQuery(() => listTeamsOptions(params));
</script>

<div class="flex flex-row items-end gap-4">
	<div class="flex flex-col h-full">
		<div class="h-fit">
			<span>Your Teams</span>
		</div>
		<div class="flex-1 flex items-end gap-4">
			<LoadingQueryWrapper query={userTeamsQuery}>
				{#snippet view(userTeams: Team[])}
					{#each userTeams as team}
						<TeamCard
							title={team.attributes.name}
							teamId={team.id}
						/>
					{/each}
				{/snippet}
			</LoadingQueryWrapper>
		</div>
	</div>
</div>
