<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import type { OncallShift } from "$lib/api";
	import LoadingQueryWrapper from "$src/components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import FilterPage from "$src/components/layout/filter-page/FilterPage.svelte";
	import ShiftCard from "$features/oncall/components/shift-card/ShiftCard.svelte";
	import RosterSelectField from "$src/components/forms/roster-select-field/RosterSelectField.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { initOncallShiftsListController } from "./controller.svelte";

	registerPageDescriptor(() => ({ title: "Oncall Shifts" }));

	const controller = initOncallShiftsListController();

	const periodTypes: string[] = [
		// PeriodType.Day,
		// PeriodType.Week,
		// PeriodType.BiWeek1,
		// PeriodType.Month,
		// PeriodType.Quarter,
		// PeriodType.CalendarYear,
	];

	// const formatShiftStatusField = (opts: MenuOption<string>[]) => {
	// 	if (opts.length === 0) return "None";
	// 	if (opts.length === statusOptions.length) return "Any";
	// 	return opts.map((o) => o.label).join(", ");
	// };
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

		<RosterSelectField onSelected={controller.onRosterSelected} />

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
	<PaginatedQueryListBox {...controller.paginatedShiftsQuery}>
		<LoadingQueryWrapper query={controller.query}>
			{#snippet view(shifts: OncallShift[])}
				{#each shifts as shift (shift.id)}
					<ShiftCard {shift} />
				{:else}
					<div class="grid place-items-center flex-1">
						<span class="text-surface-content/80">No Shifts Found</span>
					</div>
				{/each}
			{/snippet}
		</LoadingQueryWrapper>
	</PaginatedQueryListBox>
</FilterPage>
