<script lang="ts">
	import type { MeetingSchedule } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import ScheduleSessions from "./ScheduleSessions.svelte";
	import Header from "$components/header/Header.svelte";
	import Button from "$components/button/Button.svelte";

	type Props = { schedule: MeetingSchedule };
	const { schedule }: Props = $props();

	const attr = $derived(schedule.attributes);

	const timing = $derived(attr.timing);
	const repeats = $derived.by(() => {
		let rep = "week";
		if (timing.repeat === "monthly") rep = "month";
		if (timing.repeat === "daily") rep = "day";

		let label = `${timing.repeatStep} ${rep}s`;
		if (timing.repeatStep <= 1) return rep;
		if (timing.repeat === "monthly") label += ` on ${timing.repeatMonthlyOn}`;
		return label;
	});

	const untilLabel = $derived.by(() => {
		if (timing.indefinite) {
			return `indefinitely`;
		} else if (timing.untilNumRepetitions) {
			return `until ${timing.untilNumRepetitions} repetitions`;
		} else if (timing.untilDate) {
			return `until ${timing.untilDate}`;
		}
	});

	const scheduleLabel = $derived(repeats + (untilLabel ? `, ${untilLabel}` : ""));
</script>

<div class="grid grid-cols-3 gap-2 flex-1 min-h-0 overflow-y-hidden">
	<div class="flex flex-col gap-2">
		<div class="border p-2 flex items-center gap-2">
			<Avatar kind="team" id={attr.hostTeamId} />
			<span class="text-lg">host team</span>
		</div>

		<div class="border p-2 flex flex-col gap-2">
			<Header title="Schedule"></Header>
			<span>repeats every {scheduleLabel}</span>
		</div>

		<div class="border p-2 flex flex-col gap-2">
			<Header title="Attendees"></Header>
			<span>{attr.attendees.private ? "Private" : "Open to everyone"}</span>
			<div class="grid grid-cols-2">
				<div class="flex flex-col gap-2">
					{#each attr.attendees.teams as team}
						<span>{team}</span>
					{/each}

					{#each attr.attendees.users as user}
						<span>{user}</span>
					{/each}
				</div>
			</div>
		</div>
	</div>

	<div class="flex flex-col gap-2 overflow-y-auto">
		<ScheduleSessions {schedule} />
	</div>

	<div class="border p-2 overflow-y-auto">
		<Header title="Meeting Document Template">
			{#snippet actions()}
				<Button>Edit</Button>
			{/snippet}
		</Header>
	</div>
</div>
