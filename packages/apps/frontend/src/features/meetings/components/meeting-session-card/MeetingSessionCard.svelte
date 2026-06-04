<script lang="ts">
	import type { MeetingSession } from "$lib/api";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import { resolve } from "$app/paths";

	type Props = {
		session: MeetingSession;
	};
	const { session }: Props = $props();

	const scheduleId = $derived(session.attributes.meetingScheduleId);
	const start = $derived(session.attributes.startsAt);
</script>


<Card.Root class="gap-4 p-4 min-w-xs">
    <Card.Header class="p-0">
        <div class="flex items-center justify-between gap-4 h-fit">
            <div class="flex flex-col gap-2">
                <Card.Title class="capitalize">{session.attributes.title}</Card.Title>
            </div>
            <div class="flex items-center gap-2">
                <div class="px-4">
					<span>{start.toLocaleString()}</span>
				</div>
            </div>
        </div>
    </Card.Header>

	<Card.Action>
		<div class="flex justify-end px-1">
			<Button href={resolve("/meetings/sessions/[id]", session)}>View</Button>
		</div>
	</Card.Action>
</Card.Root>
