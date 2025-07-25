<script lang="ts">
	import TeamView from "$src/features/team/view/TeamView.svelte";

	let { data } = $props();
</script>

<TeamView team={data.team} />
