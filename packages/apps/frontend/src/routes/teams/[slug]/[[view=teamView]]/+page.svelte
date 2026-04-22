<script lang="ts">
	import TeamView from "$features/teams/views/team";

	let { data } = $props();
</script>

<TeamView slug={data.slug} />
