<script lang="ts">
	import type { OncallShift, User } from "$lib/api";

	type Props = {
		user: User;
		shifts: OncallShift[];
	};
	const { user, shifts }: Props = $props();

	// Calculate stats
	const totalShifts = shifts.length;
	const currentlyOncall = shifts.some(({attributes}) => {
		const now = new Date();
		const start = new Date(attributes.startAt);
		const end = new Date(attributes.endAt);
		return start <= now && end >= now;
	});

	const hoursOncall = shifts.reduce((total, shift) => {
		const start = new Date(shift.attributes.startAt);
		const end = new Date(shift.attributes.endAt);
		const hours = (end.getTime() - start.getTime()) / (1000 * 60 * 60);
		return total + hours;
	}, 0);

	const upcomingShifts = shifts
		.filter((shift) => new Date(shift.attributes.startAt) > new Date())
		.sort((a, b) => new Date(a.attributes.startAt).getTime() - new Date(b.attributes.startAt).getTime())
		.slice(0, 3);
</script>

<div class="">
	<h2>Oncall Statistics</h2>

	<div class="">
		<div class="">
			<div class="">{totalShifts}</div>
			<div class="">Total Shifts</div>
		</div>

		<div class="">
			<div class="">{Math.round(hoursOncall)}</div>
			<div class="">Hours Oncall</div>
		</div>

		<div class="">
			<div class="{currentlyOncall ? 'active' : 'inactive'}">
				{currentlyOncall ? "Yes" : "No"}
			</div>
			<div class="">Currently Oncall</div>
		</div>
	</div>

	{#if upcomingShifts.length > 0}
		<div class="">
			<h3>Upcoming Shifts</h3>
			<ul>
				{#each upcomingShifts as shift}
					{@const attr = shift.attributes}
					<li>
						<div class="">{attr.roster?.attributes.name || "Unknown Roster"}</div>
						<div class="">
							{new Date(attr.startAt).toLocaleDateString()} -
							{new Date(attr.endAt).toLocaleDateString()}
						</div>
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</div>

