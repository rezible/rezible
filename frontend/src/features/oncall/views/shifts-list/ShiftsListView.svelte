<script lang="ts">
	import { DateRangeField, TextField } from "svelte-ux";
	import { PeriodType } from "@layerstack/utils";
	import type { DateRange as DateRangeType } from "@layerstack/utils/dateRange";
	import { mdiCalendarRange, mdiMagnify } from "@mdi/js";
	import { formatDistanceToNow, subDays } from "date-fns";
	import { createQuery } from "@tanstack/svelte-query";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import { listOncallShiftsOptions, type ListOncallShiftsData, type OncallShift } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import Header from "$components/header/Header.svelte";
	import Card from "$components/card/Card.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import SearchInput from "$components/search-input/SearchInput.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Oncall Shifts", href: "/shifts" }]);

	let searchValue = $state<string>();

	let queryParams = $state<ListOncallShiftsData["query"]>({});
	const shiftsQuery = createQuery(() => listOncallShiftsOptions({query: queryParams}));

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

{#snippet filters()}
	<div class="flex flex-col gap-2">
		<SearchInput bind:value={searchValue} />

		<DateRangeField
			dense
			{periodTypes}
			bind:value={dateRange}
			on:change={(e) => {
				updateDateRange(e.detail);
			}}
			icon={mdiCalendarRange}
		/>
	</div>
{/snippet}

<FilterPage {filters}>
	<div class="flex flex-col gap-2 min-h-0 flex-1 overflow-auto">
		<LoadingQueryWrapper query={shiftsQuery}>
			{#snippet view(shifts: OncallShift[])}
				{#each shifts as shift}
					{@const attr = shift.attributes}
					{@const roster = attr.roster.attributes}
					<a href="/shifts/{shift.id}">
						<Card classes={{ root: "w-full hover:bg-primary/30", headerContainer: "py-2" }}>
							{#snippet header()}
								<Header title={roster.name} subheading={attr.role} />
							{/snippet}

							{#snippet contents()}
								<div class="pb-2">
									<span>{formatDistanceToNow(attr.endAt)} ago</span>
								</div>
							{/snippet}
						</Card>
					</a>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</FilterPage>
