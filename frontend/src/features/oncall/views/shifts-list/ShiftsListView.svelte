<script lang="ts">
	import { DateRangeField, MultiSelectField, type MenuOption } from "svelte-ux";
	import { PeriodType } from "@layerstack/utils";
	import type { DateRange as DateRangeType } from "@layerstack/utils/dateRange";
	import { mdiCalendarRange } from "@mdi/js";
	import { subDays } from "date-fns";
	import { createQuery } from "@tanstack/svelte-query";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import { listOncallShiftsOptions, type ListOncallShiftsData, type OncallShift } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import ShiftCard from "$features/oncall/components/shift-card/ShiftCard.svelte";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import RosterSelectField from "$src/components/roster-select-field/RosterSelectField.svelte";
	import { watch } from "runed";
	import PaginatedListBox from "$src/components/paginated-listbox/PaginatedListBox.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Oncall Shifts", href: "/shifts" }]);

	const statusOptions: MenuOption<string>[] = [
		{ label: 'Active', value: "active" },
		{ label: 'Past', value: "past" },
		{ label: 'Upcoming', value: "upcoming", disabled: true },
	];
	let selectedStatus = $state<string[]>(statusOptions.map(o => o.value));

	const pagination = createPaginationStore();

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

	const onRosterSelected = (id?: string) => {
		if (!id) return;
	}

	const formatShiftStatusField = (opts: MenuOption<string>[]) => {
		if (opts.length === 0) return "None";
		if (opts.length === statusOptions.length) return "Any";
		return opts.map((o) => o.label).join(", ");
	};

	const setShiftStatus = (value?: string[]) => {
		if (!value || value.length === 0) return;
		selectedStatus = value;
	};

	const queryParams = $derived<ListOncallShiftsData["query"]>({});
	const shiftsQuery = createQuery(() => listOncallShiftsOptions({query: queryParams}));
	const queryPagination = $derived(shiftsQuery.data?.pagination);
	watch(() => queryPagination, p => {
		if (!p) return;
		pagination.setTotal(p.total)
	})
</script>

{#snippet filters()}
	<div class="flex flex-col gap-2">
		<MultiSelectField
			label="Shift Status"
			labelPlacement="top"
  			formatSelected={c => formatShiftStatusField(c.options)}
			options={statusOptions}
			bind:value={() => selectedStatus, setShiftStatus}
			clearable={false}
			mode="actions"
			maintainOrder
		>
			<div slot="actions" let:selection class="flex items-center">
				{#if !selection.selected || (Array.isArray(selection.selected) && selection.selected.length === 0)}
					<div class="text-sm text-danger">Nothing selected</div>
				{/if}
			</div>
		</MultiSelectField>

		<RosterSelectField onSelected={onRosterSelected} />

		<DateRangeField
			label="Date Range"
			labelPlacement="top"
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
	<PaginatedListBox {pagination}>
		<LoadingQueryWrapper query={shiftsQuery}>
			{#snippet view(shifts: OncallShift[])}
				{#each shifts as shift}
					<ShiftCard {shift} />
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
