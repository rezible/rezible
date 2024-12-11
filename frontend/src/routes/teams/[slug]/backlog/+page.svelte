<script lang="ts">
    import { createQuery } from '@tanstack/svelte-query';
    import type { Team } from '$lib/api/index.js';
    import LoadingQueryWrapper from '$components/loader/LoadingQueryWrapper.svelte';
	import TeamBacklogView from "$features/teams/views/team/team-backlog/TeamBacklogView.svelte";

	let { data } = $props();

	const query = createQuery(() => data.teamQueryOptions);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(team: Team)}
		<TeamBacklogView {team} />
	{/snippet}
</LoadingQueryWrapper>