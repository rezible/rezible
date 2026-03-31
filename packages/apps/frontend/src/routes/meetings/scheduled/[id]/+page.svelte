<script lang="ts">
	import { page } from "$app/state";
	import MeetingScheduleView from "$features/meetings/views/meeting-schedule";
</script>

<MeetingScheduleView id={page.params.id || ""} />