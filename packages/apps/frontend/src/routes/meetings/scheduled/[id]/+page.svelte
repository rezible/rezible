<script lang="ts">
	import type { PageProps } from "./$types";
	import MeetingScheduleView from "$features/meetings/views/meeting-schedule";

	const { params }: PageProps = $props();
</script>

<MeetingScheduleView id={params.id} />
