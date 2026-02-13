<script lang="ts">
	import { page } from "$app/state";
	import { MeetingSessionView, setMeetingSessionViewState } from "$features/meeting-session";

	setMeetingSessionViewState(() => (page.params.id || ""));
</script>

<MeetingSessionView />
