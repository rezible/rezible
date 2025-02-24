<script lang="ts">
	import MeetingsOverview from "$features/meetings/views/meetings-overview/MeetingsOverview.svelte";
	import { setPageBreadcrumbs } from "$lib/appState.svelte";

	setPageBreadcrumbs(() => [{ label: "Meetings" }]);
</script>

<MeetingsOverview />
