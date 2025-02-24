<script lang="ts">
	import TeamBacklogView from "$features/teams/views/team/team-backlog/TeamBacklogView.svelte";

	let { data } = $props();
</script>

<TeamBacklogView teamId={data.id} />
