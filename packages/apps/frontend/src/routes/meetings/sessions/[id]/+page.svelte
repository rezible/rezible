<script lang="ts">
	import type { PageProps } from "./$types";
	import MeetingSessionView from "$features/meetings/views/meeting-session";

	const { params }: PageProps = $props();
</script>

<MeetingSessionView id={params.id} />
