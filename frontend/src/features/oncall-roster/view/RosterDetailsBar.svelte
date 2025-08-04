<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";
	import ShiftProgressCircle from "$features/oncall-shifts-list/components/shift-card/ShiftProgressCircle.svelte";
	import { useOncallRosterViewState } from "$features/oncall-roster";
	
	const view = useOncallRosterViewState();

	// TODO: include this
	const teamId = $derived(view.rosterId ?? "");
	const shift = $derived(view.activeShift);

	const userLocalTime = new Date(Date.now()).toLocaleTimeString(undefined, {hour: "2-digit", minute: "2-digit", hour12: true})
</script>

{#if shift}
	<a href="/shifts/{shift.id}" class="flex items-center gap-4 px-4 bg-success-900/50 rounded-lg hover:bg-success-900/40">
		<div class="flex flex-col">
			<span class="text-xs">Currently Oncall</span>
			<div class="flex items-center align-middle gap-2">
				<Avatar kind="user" size={14} id={shift.id} />
				<span class="text-sm font-semibold">{shift.attributes.user.attributes.name}</span>
				<span class="text-xs text-surface-content/70 align-middle">({userLocalTime})</span>
			</div>
		</div>
		<div class="">
			<ShiftProgressCircle {shift} size={30} pulse={false} />
		</div>
	</a>
{/if}

<a href="/teams/{teamId}" class="flex items-center gap-4 px-4 bg-accent-900/50 rounded-lg hover:bg-accent-900/40 h-full">
	<div class="flex flex-col">
		<span class="text-xs">Team</span>
		<div class="flex items-center gap-2">
			<Avatar kind="team" size={14} id={teamId} />
			<span class="text-sm font-semibold">Test Team</span>
		</div>
	</div>
</a>