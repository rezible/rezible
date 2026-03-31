<script lang="ts">
	import { page } from "$app/state";
	import MeetingSessionView from "$features/meetings/views/meeting-session";
</script>

<MeetingSessionView id={page.params.id || ""} />
