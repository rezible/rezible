<script lang="ts">
    import type { MeetingSession } from "$lib/api";
    import { addMinutes } from "date-fns";
    import { Button, Card } from "svelte-ux";

	type Props = {
		session: MeetingSession;
	}	
	const { session }: Props = $props();

	const scheduleId = $derived(session.attributes.meeting_schedule_id);
	const start = $derived(session.attributes.starts_at)
</script>

<a href="/meetings/{scheduleId}/{session.id}">
	<Card title={session.attributes.title} class="w-full">
		<div class="px-4">
			<span>{start.toLocaleString()}</span>
		</div>

		<div slot="actions" class="flex justify-end px-1">
			<Button>View</Button>
		</div>
	</Card>
</a>