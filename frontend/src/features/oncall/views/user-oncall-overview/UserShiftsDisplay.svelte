<script lang="ts">
	import { Button, Header, ListItem, Icon, Month, MonthList } from "svelte-ux";
	import type { DateRange } from "@layerstack/utils/dateRange";
	import { startOfWeek, endOfWeek } from "date-fns";
	import type { OncallShift, UserOncallDetails } from "$lib/api";
	import ActiveShiftCard from "./ActiveShiftCard.svelte";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { mdiChevronRight } from "@mdi/js";
	import { formatDuration, minutesToHours, differenceInMinutes } from "date-fns";

	type Props = {
		details: UserOncallDetails;
		// activeShifts: OncallShift[];
		// upcomingShifts: OncallShift[];
		// pastShifts: OncallShift[];
	}
	// let { activeShifts, upcomingShifts, pastShifts }: Props = $props();
	const { details }: Props = $props();

	const { activeShifts, upcomingShifts, pastShifts } = $derived(details);
	// const activeShifts = $derived(details.activeShifts);
	const numActive = $derived(activeShifts.length);
	const activeShiftsSubheading = $derived(
		`You are currently oncall for ${numActive} roster${numActive > 1 ? "s" : ""}`
	);

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

	let selectedWeek = $state<DateRange>({ from: startOfWeek(Date.now()), to: endOfWeek(Date.now()) });

	const onMonthDateChange = (e: CustomEvent<Date>) => {
		const date = e.detail;
		selectedWeek = { from: startOfWeek(date), to: endOfWeek(date) };
	};
</script>

<div class="flex flex-col gap-2 min-h-0">
	{#if numActive > 0}
		<div class="flex flex-col col-span-2 border rounded-lg p-2">
			<Header title="Active" subheading={activeShiftsSubheading} />

			<div class="w-full h-0 border-b mt-1 mb-2"></div>

			<div class="flex flex-row overflow-x-auto gap-2">
				{#each activeShifts as shift}
					<ActiveShiftCard {shift} />
				{/each}
			</div>
		</div>
	{/if}

	<!--div class="h-full grid grid-cols-5 gap-2">
		<div class="grid col-span-1">
			<MonthList />
		</div>
		<div class="pb-2 border col-span-4">
			<Month on:dateChange={onMonthDateChange} showOutsideDays />
		</div>
	</div-->
	<div class="flex-1 min-h-0 grid grid-cols-2 gap-2">
		<div class="flex flex-col min-h-0 border rounded-lg p-2">
			<Header title="Past" subheading="Last 30 days" />

			<div class="w-full h-0 border-b my-2"></div>

			<div class="flex flex-col gap-2 flex-1 overflow-auto">
				{#each pastShifts as shift}
					{@render shiftListItem(shift)}
				{/each}
			</div>
		</div>

		<div class="flex flex-col min-h-0 border rounded-lg p-2">
			<Header title="Upcoming" subheading="Next 7 days" />

			<div class="w-full h-0 border-b my-2"></div>

			<div class="flex flex-col gap-2 flex-1 overflow-y-auto">
				{#each upcomingShifts as shift}
					{@render shiftListItem(shift)}
				{/each}
			</div>
		</div>
	</div>
</div>

{#snippet shiftListItem(shift: OncallShift)}
	{@const roster = shift.attributes.roster}
	{@const duration = formatShiftDuration(shift)}
	<a href="/oncall/shifts/{shift.id}">
		<ListItem title={roster.attributes.name} classes={{ root: "hover:bg-secondary-900" }}>
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
		</ListItem>
	</a>
{/snippet}
