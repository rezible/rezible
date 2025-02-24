<script lang="ts">
	import { page } from "$app/state";
	import ScheduledMeetingView from "$src/features/meetings/views/scheduled-meeting/ScheduledMeetingView.svelte";
</script>

<ScheduledMeetingView scheduleId={page.params.id} />