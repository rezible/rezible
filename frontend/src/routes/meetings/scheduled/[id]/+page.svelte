<script lang="ts">
	import { page } from "$app/state";
	import { MeetingScheduleView, setMeetingScheduleViewState } from "$features/meeting-schedule";

	setMeetingScheduleViewState(() => (page.params.id || ""));
</script>

<MeetingScheduleView />