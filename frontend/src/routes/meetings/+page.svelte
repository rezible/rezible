<script lang="ts">
	import MeetingsOverview from "$features/meetings/views/meetings-overview/MeetingsOverview.svelte";
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";

	setPageBreadcrumbs(() => [{ label: "Meetings" }]);
</script>

<MeetingsOverview />
