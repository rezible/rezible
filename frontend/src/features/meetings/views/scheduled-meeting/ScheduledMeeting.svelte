<script lang="ts">
	import type { MeetingSchedule, MeetingScheduleAttributes, MeetingScheduleTiming } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import ScheduleSessions from "./ScheduleSessions.svelte";

	type Props = { schedule: MeetingSchedule };
	const { schedule }: Props = $props();

	const attr = $derived(schedule.attributes);

	const getRepeats = (s: MeetingScheduleTiming["repeat"], step: number) => {
		let rep = "week";
		if (s === "monthly") rep = "month";
		if (s === "daily") rep = "day";

		if (step <= 1) return rep;
		return `${step} ${rep}s`;
	};
</script>

<div class="grid grid-cols-3 gap-2 flex-1 min-h-0 overflow-y-hidden">
	{@render scheduleAttributesView(attr)}

	<div class="border p-2 flex flex-col gap-2 overflow-y-auto">
		<ScheduleSessions {schedule} />
	</div>

	<div class="border p-2 overflow-y-auto">
		<div class="h-32">meeting document template</div>
	</div>
</div>

{#snippet timingView(t: MeetingScheduleTiming)}
	<span>repeats every {getRepeats(t.repeat, t.repeatStep)}</span>
	{#if t.repeat === "monthly"}
		<span>on {t.repeatMonthlyOn}</span>
	{/if}
	{#if t.indefinite}
		<span>indefinitely</span>
	{:else if t.untilNumRepetitions}
		<span>until {t.untilNumRepetitions} repetitions</span>
	{:else if t.untilDate}
		<span>until {t.untilDate}</span>
	{/if}
{/snippet}

{#snippet scheduleAttributesView(attr: MeetingScheduleAttributes)}
	<div class="flex flex-col gap-2">
		<div class="border p-2 flex items-center gap-2">
			<Avatar kind="team" id={attr.hostTeamId} />
			<span class="text-lg">host team</span>
		</div>

		<div class="border p-2 flex flex-col gap-2">
			{@render timingView(attr.timing)}
		</div>

		<div class="border p-2 flex flex-col gap-2">
			<span class="text-lg">Invites</span>
			<span>{attr.attendees.private ? "Private" : "Open to everyone"}</span>
			<span>Users: {JSON.stringify(attr.attendees.users)}</span>
			<span>Teams: {JSON.stringify(attr.attendees.teams)}</span>
		</div>
	</div>
{/snippet}
