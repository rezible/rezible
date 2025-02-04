<script lang="ts">
	import type { MeetingSchedule, MeetingScheduleTiming } from "$lib/api";
	import { addMinutes } from "date-fns";
	import { Button, Card } from "svelte-ux";

	type Props = {
		schedule: MeetingSchedule;
	};
	const { schedule }: Props = $props();

	const attr = $derived(schedule.attributes);
	const href = `/meetings/scheduled/${schedule.id}`;

	const getNextScheduled = (m: MeetingScheduleTiming) => {
		return addMinutes(Date.now(), 60);
	};
</script>

<Card title={attr.name} class="w-full">
	<div class="px-4">
		<span>{getNextScheduled(attr.timing).toLocaleString()}</span>
	</div>

	<div slot="actions" class="flex justify-end px-1">
		<Button {href}>Edit</Button>
		<Button {href}>View</Button>
	</div>
</Card>
