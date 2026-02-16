<script lang="ts">
	import { mdiChevronRight, mdiFire, mdiPhone } from "@mdi/js";
	import { formatDuration, minutesToHours, differenceInMinutes } from "date-fns";
	import type { OncallShift } from "$lib/api";
	import MetricCard from "$components/viz/MetricCard.svelte";
	import Header from "$components/header/Header.svelte";
	import { useUserViewController } from "./controller.svelte";

	const view = useUserViewController();
	const shifts = $derived(view.oncallShifts ?? []);

	// Calculate stats
	const totalShifts = $derived(shifts.length);

	// TODO: make this a component (currently taken from user oncall view)

	const formatShiftDuration = (shift: OncallShift) => {
		const start = new Date(shift.attributes.startAt);
		const end = new Date(shift.attributes.endAt);
		const minutes = differenceInMinutes(end, start);
		if (minutes < 60) return `${minutes} minutes`;
		const hours = minutesToHours(minutes);
		const remainingMinutes = minutes - hours * 60;
		if (hours < 24)
			return formatDuration({ hours, minutes: remainingMinutes }, { format: ["hours", "minutes"] });
		const days = Math.floor(hours / 24);
		const remainingHours = hours - days * 24;
		return formatDuration(
			{ days, hours: remainingHours, minutes: remainingMinutes },
			{ format: ["days", "hours", "minutes"] }
		);
	};
</script>

<div class="flex flex-col gap-1 p-2">
	<Header title="Oncall" classes={{title: "text-2xl", root: ""}} />

	<div class="flex gap-2 flex-wrap">
		<MetricCard title="Total Shifts" icon={mdiPhone} metric={totalShifts} />
		<MetricCard title="Total Incidents" icon={mdiFire} metric={totalShifts} />
	</div>

	<div class="w-full h-0 border-b mt-2 mb-1"></div>

	<div class="flex flex-col">
		<Header title="Recent Shifts" classes={{title: "text-xl"}} />
		
		<div class="flex flex-col gap-2">
			{#each shifts as shift}
				{@render shiftListItem(shift)}
			{/each}
		</div>
	</div>
</div>


{#snippet shiftListItem(shift: OncallShift)}
	{@const roster = shift.attributes.roster}
	{@const duration = formatShiftDuration(shift)}
	<a href="/shifts/{shift.id}">
		<span>shift for {roster.attributes.name}</span>
		<!-- <ListItem title={roster.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
			<svelte:fragment slot="avatar">
				<Avatar kind="roster" size={32} id={roster.id} />
			</svelte:fragment>
			<svelte:fragment slot="subheading">
				<span class="text-surface-content"
					><span class="font-bold">{shift.attributes.role}</span> for {duration}</span
				>
			</svelte:fragment>
			<div slot="actions">
				<Icon data={mdiChevronRight} />
			</div>
		</ListItem> -->
	</a>
{/snippet}
