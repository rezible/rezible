<script lang="ts">
	import { Field, ToggleGroup, ToggleOption } from "svelte-ux";
	import Header from "$components/header/Header.svelte";
	import { Button, DateRangeField, Pagination } from "svelte-ux";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventsFilters from "$components/oncall-events/EventsFilters.svelte";
	import EventRow from "$components/oncall-events/EventRow.svelte";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import { PeriodType } from "@layerstack/utils";
	import { mdiCalendarRange, mdiFilter } from "@mdi/js";
	import { AnnotationDialogState, setAnnotationDialogState } from "$components/oncall-events/annotation-dialog/dialogState.svelte";
	import { OncallEventsTableState, type DateRangeOption } from "./eventsTable.svelte";
	import type { OncallAnnotation } from "$src/lib/api";
	import { watch } from "runed";

	const tableState = new OncallEventsTableState();

	let filtersVisible = $state(false);

	setAnnotationDialogState(new AnnotationDialogState({
		onClosed: (updated?: OncallAnnotation) => {
			if (updated) tableState.invalidateQuery();
		},
	}));

	const dateRangeOptions: DateRangeOption[] = [
		{label: "Last 7 Days", value: "7d"},
		{label: "Last Month", value: "30d"}, 
		{label: "Custom", value: "custom"},
	];

	watch(() => tableState.dateRangeOption, opt => {
		if (opt === "custom" && !filtersVisible) filtersVisible = true;
	})
</script>

<EventAnnotationDialog />

<div class="w-full h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Oncall Events" classes={{root: "p-2 w-full", title: "text-xl"}}>
		{#snippet actions()}
			<div class="justify-end flex gap-2 items-end">
				<Field dense base classes={{root: "", container: "px-0 border-none py-0", input: "my-0 gap-2"}}>
					<ToggleGroup variant="outline" inset classes={{root: "bg-surface-100"}} bind:value={tableState.dateRangeOption}>
						{#if !!tableState.activeShift}
							<ToggleOption value="shift">Active Shift</ToggleOption>
						{/if}

						{#each dateRangeOptions as opt}
							<ToggleOption value={opt.value}>{opt.label}</ToggleOption>
						{/each}
					</ToggleGroup>
				</Field>

				<Button icon={mdiFilter} iconOnly 
					variant={filtersVisible ? "fill-light" : "default"}
					color={filtersVisible ? "accent" : "default"}
					on:click={() => {filtersVisible = !filtersVisible}} 
				/>
			</div>
		{/snippet}
	</Header>

	{#if filtersVisible}
		<div class="w-full p-2 pt-0">
			<EventsFilters bind:filters={tableState.filters}>
				{#snippet extra()}
					{#if tableState.dateRangeOption === "custom"}
						<DateRangeField
							label="Custom Date Range"
							periodTypes={[PeriodType.Day]}
							getPeriodTypePresets={() => []}
							dense
							classes={{
								field: { root: "gap-0", container: "pl-0 py-[2px] flex items-center", prepend: "[&>span]:mr-2" },
							}}
							icon={mdiCalendarRange}
							bind:value={() => tableState.dateRange, d => (tableState.customDateRangeValue = d)}
						/>
					{/if}
				{/snippet}
			</EventsFilters>
		</div>
	{/if}

	<div class="flex-1 flex flex-col overflow-y-auto border-t">
		{#if tableState.loading}
			<LoadingIndicator />
		{:else}
			{#each tableState.pageData as event (event.id)}
				<EventRow {event} />
			{:else}
				<div class="grid place-items-center flex-1">
					<span class="text-surface-content/80">No Events</span>
				</div>
			{/each}
		{/if}
	</div>

	{#if tableState.pagination.current.totalPages > 0}
		<Pagination
			pagination={tableState.paginationStore}
			perPageOptions={[10, 25, 50]}
			show={["perPage", "pagination", "prevPage", "nextPage"]}
			classes={{
				root: "border-t py-1",
				perPage: "flex-1 text-right",
				pagination: "px-8",
			}}
		/>
	{/if}
</div>