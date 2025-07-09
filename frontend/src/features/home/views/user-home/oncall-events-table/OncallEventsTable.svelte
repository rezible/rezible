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
</script>

<EventAnnotationDialog />

<div class="w-full h-full max-h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Oncall Events" subheading="Operational events during these dates" classes={{root: "p-2 w-full", title: "text-xl"}}>
		{#snippet actions()}
			<div class="justify-end flex gap-2 items-end">
				<Field label="Date Range" labelPlacement="top" dense base classes={{root: "", container: "px-0 border-none py-0", input: "my-0 gap-2"}}>
					<ToggleGroup bind:value={tableState.dateRangeOption} variant="outline" inset classes={{root: "bg-surface-100"}}>
						{#if !!tableState.activeShift}
							<ToggleOption value="shift">Active Shift</ToggleOption>
						{/if}

						{#each dateRangeOptions as opt}
							<ToggleOption value={opt.value}>{opt.label}</ToggleOption>
						{/each}
					</ToggleGroup>

					{#if tableState.dateRangeOption === "custom"}
						<DateRangeField
							label=""
							labelPlacement="top"
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
			<EventsFilters bind:filters={tableState.filters} />
		</div>
	{/if}

	<div class="flex flex-col flex-1 overflow-y-auto border-t">
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