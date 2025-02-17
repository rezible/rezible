<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, Card, Collapse, Header } from "svelte-ux";
	import { mdiFilter } from "@mdi/js";
	import { listMeetingSessionsOptions, type ListMeetingSessionsData, type MeetingSession } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import MeetingSessionCard from "$features/meetings/components/meeting-session-card/MeetingSessionCard.svelte";

	let queryParams = $state<ListMeetingSessionsData["query"]>({});
	const query = createQuery(() => listMeetingSessionsOptions({ query: queryParams }));
</script>

<div class="flex flex-col gap-2 min-h-0 h-full">
	<Collapse>
		<div slot="trigger" class="flex-1">
			<Button icon={mdiFilter} iconOnly />
			Filters
		</div>
		<div class="px-3 pb-3 border-t">
			show filters
		</div>
	</Collapse>

	<div class="flex-1 flex flex-col gap-2 overflow-y-auto">
		<LoadingQueryWrapper {query}>
			{#snippet view(sessions: MeetingSession[])}
				{#each sessions as session}
					<MeetingSessionCard {session} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
