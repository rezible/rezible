<script lang="ts">
	import type { MeetingSession } from "$lib/api";
	import Button from "$components/button/Button.svelte";
	import Card from "$components/card/Card.svelte";
	import Header from "$components/header/Header.svelte";

	type Props = {
		session: MeetingSession;
	};
	const { session }: Props = $props();

	const scheduleId = $derived(session.attributes.meetingScheduleId);
	const start = $derived(session.attributes.startsAt);
</script>

<a href="/meetings/{scheduleId}/{session.id}">
	<Card classes={{root: "w-full"}}>
		{#snippet header()}
			<Header title={session.attributes.title} />
		{/snippet}

		<div class="px-4">
			<span>{start.toLocaleString()}</span>
		</div>

		{#snippet actions()}
			<div class="flex justify-end px-1">
				<Button>View</Button>
			</div>
		{/snippet}
	</Card>
</a>
