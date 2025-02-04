<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import type { Team } from "$lib/api/index.js";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import TeamView from "$features/teams/views/team/TeamView.svelte";

	let { data } = $props();

	const query = createQuery(() => data.teamQueryOptions);
</script>

<LoadingQueryWrapper {query}>
	{#snippet view(team: Team)}
		<TeamView {team} />
	{/snippet}
</LoadingQueryWrapper>
