<script lang="ts">
	import { watch } from "runed";
	import { mdiFilter } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import Header from "$components/header/Header.svelte";
	import EventAnnotationDialog from "$components/events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventRow from "$components/events/EventRow.svelte";
	import LoadingIndicator from "$components/loading-indicator/LoadingIndicator.svelte";
	import { AnnotationDialogState, setAnnotationDialogState } from "$components/events/annotation-dialog/dialogState.svelte";
	import { EventsTableState } from "./eventsTableState.svelte";
	import type { EventAnnotation } from "$lib/api";
	import EventsFilters from "./EventsFilters.svelte";

	const tableState = new EventsTableState();

	let filtersVisible = $state(false);

	setAnnotationDialogState(new AnnotationDialogState({
		onClosed: (updated?: EventAnnotation) => {
			if (updated) tableState.invalidateQuery();
		},
	}));

	watch(() => tableState.dateRangeOption, opt => {
		if (opt === "custom" && !filtersVisible) filtersVisible = true;
	});
</script>

<EventAnnotationDialog />

<div class="w-full h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Events" subheading="Recent oncall events" classes={{root: "p-2 w-full", title: "text-xl"}}>
		{#snippet actions()}
			<div class="justify-end flex gap-2 items-end">
				<!--Field dense base classes={{root: "", container: "px-0 border-none py-0", input: "my-0 gap-2"}}>
					<ToggleGroup variant="outline" inset classes={{root: "bg-surface-100"}} bind:value={tableState.dateRangeOption}>
						{#if !!tableState.activeShift}
							<ToggleOption value="shift">Active Shift</ToggleOption>
						{/if}

						{#each dateRangeOptions as opt}
							<ToggleOption value={opt.value}>{opt.label}</ToggleOption>
						{/each}
					</ToggleGroup>
				</Field-->

				<Button 
					color={filtersVisible ? "accent" : "default"}
					onclick={() => {filtersVisible = !filtersVisible}} 
				/>
			</div>
		{/snippet}
	</Header>

	{#if filtersVisible}
		<div class="w-full p-2 pt-0">
			<EventsFilters {tableState} />
		</div>
	{/if}

	<div class="flex-1 flex flex-col overflow-y-auto border-t">
		{#if tableState.loading}
			<LoadingIndicator />
		{:else}
			{#each tableState.events as event (event.id)}
				<EventRow {event} />
			{:else}
				<div class="grid place-items-center flex-1">
					<span class="text-surface-content/80">No Events</span>
				</div>
			{/each}
		{/if}
	</div>

	<!-- <Pagination {...tableState.paginator.paginationProps}
		classes={{
			root: "border-t py-1",
			perPage: "flex-1 text-right",
			pagination: "px-8",
		}}
	/> -->
</div>