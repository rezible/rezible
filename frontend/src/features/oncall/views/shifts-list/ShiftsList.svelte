<script lang="ts">
	import { DateRangeField, TextField } from "svelte-ux";
	import { PeriodType } from "@layerstack/utils";
	import type { DateRange as DateRangeType } from "@layerstack/utils/dateRange";
	import { mdiCalendarRange, mdiMagnify } from "@mdi/js";
	import { formatDistanceToNow, subDays } from "date-fns";
	import { createQuery } from "@tanstack/svelte-query";
	import { listOncallShiftsOptions, type ListOncallShiftsData, type OncallShift } from "$lib/api";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import Header from "$src/components/header/Header.svelte";
	import Card from "$src/components/card/Card.svelte";

	type Props = {};
	const {}: Props = $props();

	let params = $state<ListOncallShiftsData>();
	const shiftsQuery = createQuery(() => listOncallShiftsOptions(params));

	const today = new Date();
	let dateRange = $state<DateRangeType>({
		from: subDays(today, 3),
		to: today,
		periodType: PeriodType.Day,
	});

	const periodTypes: PeriodType[] = [
		PeriodType.Day,
		PeriodType.Week,
		PeriodType.BiWeek1,
		PeriodType.Month,
		PeriodType.Quarter,
		PeriodType.CalendarYear,
	];

	const updateDateRange = (newRange: DateRangeType) => {
		console.log(newRange);
	};
</script>

<div class="flex flex-col gap-2 overflow-x-hidden">
	<div class="grid grid-cols-2 gap-2">
		<DateRangeField
			dense
			{periodTypes}
			bind:value={dateRange}
			on:change={(e) => {
				updateDateRange(e.detail);
			}}
			icon={mdiCalendarRange}
		/>

		<TextField
			label="Search"
			dense
			on:change={(e) => console.log(e.detail)}
			debounceChange
			iconRight={mdiMagnify}
			labelPlacement="float"
		/>
	</div>

	<div class="w-full border-b"></div>

	<div class="flex flex-col gap-2 min-h-0 flex-1 overflow-auto p-1">
		<LoadingQueryWrapper query={shiftsQuery}>
			{#snippet view(shifts: OncallShift[])}
				{#each shifts as shift}
					{@const attr = shift.attributes}
					{@const roster = attr.roster.attributes}
					<Card classes={{ root: "w-full", headerContainer: "py-2" }}>
						{#snippet header()}
							<Header title={roster.name} subheading={attr.role} />
						{/snippet}

						{#snippet contents()}
							<div class="pb-2">
								<span>{formatDistanceToNow(attr.endAt)} ago</span>
							</div>
						{/snippet}
					</Card>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>
