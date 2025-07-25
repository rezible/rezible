<script lang="ts">
	import TeamView from "$features/team/view/TeamView.svelte";

	let { data } = $props();
</script>

<TeamView team={data.team} />
