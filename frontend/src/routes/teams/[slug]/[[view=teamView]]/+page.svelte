<script lang="ts">
	import TeamView from "$features/teams/views/team/TeamView.svelte";

	let { data } = $props();
</script>

<TeamView id={data.team.id} />
