<script lang="ts">
	import { page } from "$app/state";
	import MeetingSessionView from "$features/meetings/views/meeting-session/MeetingSessionView.svelte";
</script>

<MeetingSessionView sessionId={page.params.id} />
