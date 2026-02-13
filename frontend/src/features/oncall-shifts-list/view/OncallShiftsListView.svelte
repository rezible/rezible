<script lang="ts">
	import { mdiCalendarRange } from "@mdi/js";
	import { subDays } from "date-fns";
	import { createQuery } from "@tanstack/svelte-query";
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { listOncallShiftsOptions, type ListOncallShiftsData, type OncallShift } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import FilterPage from "$components/filter-page/FilterPage.svelte";
	import ShiftCard from "$features/oncall-shifts-list/components/shift-card/ShiftCard.svelte";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import RosterSelectField from "$components/roster-select-field/RosterSelectField.svelte";
	import { watch } from "runed";
	import PaginatedListBox from "$components/paginated-listbox/PaginatedListBox.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Oncall Shifts", href: "/shifts" }]);

	const statusOptions = [
		{ label: 'Active', value: "active" },
		{ label: 'Past', value: "past" },
		{ label: 'Upcoming', value: "upcoming", disabled: true },
	];
	let selectedStatus = $state<string[]>(statusOptions.map(o => o.value));

	const pagination = createPaginationStore();

	const today = new Date();
	let dateRange = $state({
		from: subDays(today, 3),
		to: today,
		periodType: "day",
	});

	const periodTypes: string[] = [
		// PeriodType.Day,
		// PeriodType.Week,
		// PeriodType.BiWeek1,
		// PeriodType.Month,
		// PeriodType.Quarter,
		// PeriodType.CalendarYear,
	];

	const updateDateRange = (newRange: any) => {
		console.log(newRange);
	};

	const onRosterSelected = (id?: string) => {
		if (!id) return;
	}

	// const formatShiftStatusField = (opts: MenuOption<string>[]) => {
	// 	if (opts.length === 0) return "None";
	// 	if (opts.length === statusOptions.length) return "Any";
	// 	return opts.map((o) => o.label).join(", ");
	// };

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
		<span>status select</span>
		<!--MultiSelectField
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
		</MultiSelectField-->

		<RosterSelectField onSelected={onRosterSelected} />

		<span>date range</span>
		<!--DateRangeField
			label="Date Range"
			labelPlacement="top"
			{periodTypes}
			bind:value={dateRange}
			on:change={(e) => {
				updateDateRange(e.detail);
			}}
			icon={mdiCalendarRange}
		/-->
	</div>
{/snippet}

<FilterPage {filters}>
	<PaginatedListBox {pagination}>
		<LoadingQueryWrapper query={shiftsQuery}>
			{#snippet view(shifts: OncallShift[])}
				{#each shifts as shift}
					<ShiftCard {shift} />
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Shifts Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedListBox>
</FilterPage>
