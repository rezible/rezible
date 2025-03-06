<script lang="ts">
	import TeamView from "$features/teams/views/team/TeamView.svelte";

	let { data } = $props();
</script>

<TeamView teamSlug={data.slug} />
