<script lang="ts">
	import { TeamView, setTeamViewState } from "$features/team";

	let { data } = $props();

	setTeamViewState(() => data.slug);
</script>

<TeamView />
