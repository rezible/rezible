<script lang="ts">
	import { listOncallEventsOptions, type OncallShift } from "$lib/api";
	import { Button, Checkbox, Menu, MenuItem, Pagination, Toggle } from "svelte-ux";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { fromStore } from "svelte/store";
	import EventsFilters from "./EventsFilters.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiDotsVertical } from "@mdi/js";
	import type { DateRange } from "@layerstack/utils/dateRange";
	import { subDays } from "date-fns";
	import { PeriodType } from "@layerstack/utils";

	type Props = {
		activeShift?: OncallShift;
	};
	const { activeShift }: Props = $props();

	const defaultPerPage = 10;
	const paginationStore = createPaginationStore({ perPage: defaultPerPage });
	const pagination = fromStore(paginationStore);

	type Filters = {rosters: string[], actionRequired: boolean, dateRange: DateRange};

	const today = new Date();
	const filters = $state<Filters>({
		rosters: [],
		actionRequired: false,
		dateRange: {from: subDays(today, 7), to: today, periodType: PeriodType.Day}
	});

	const eventsQueryOpts = $derived(listOncallEventsOptions({ query: { rosterId: filters.rosters } }));
	const eventsQuery = createQuery(() => eventsQueryOpts);
	const eventsData = $derived(eventsQuery.data?.data ?? []);

	const totalEvents = $derived(eventsData.length);
	$effect(() => paginationStore.setTotal(totalEvents));

	const pageData = $derived(pagination.current.slice(eventsData));
</script>

<div class="w-full h-full max-h-full overflow-y-auto border rounded-lg flex flex-col">
	<div class="w-full flex items-center p-2">
		<div class="flex-1 px-2">
			<span class="text-2xl">Events</span>
		</div>
		<div class="">
			<EventsFilters bind:rosters={filters.rosters} bind:actionRequired={filters.actionRequired} bind:dateRange={filters.dateRange} />
		</div>
	</div>

	<div class="min-h-0 flex flex-col overflow-y-auto">
		<div class="grid grid-flow-row grid-cols-[64px_100px_minmax(100px,1fr)_64px] min-h-0 overflow-y-auto">
			<div class="grid grid-cols-subgrid col-span-full sticky top-0 bg-surface-100 items-center py-2">
				<div class="grid place-items-center"><Checkbox /></div>
				<span>Kind</span>
				<span>Title</span>
				<span class="justify-self-end pr-2">Actions</span>
			</div>

			{#each pageData ?? [] as row}
				<div class="grid grid-cols-subgrid col-span-full hover:bg-surface-100/50 py-1">
					<div class="grid place-items-center"><Checkbox /></div>
					<div class="flex items-center"><span>{row.kind}</span></div>
					<div><span>{row.title}</span></div>
					<div class="grid place-items-center">
						<Toggle let:on={open} let:toggle let:toggleOff>
							<Button icon={mdiDotsVertical} iconOnly size="sm" on:click={toggle}>
								<Menu {open} on:close={toggleOff} placement="bottom-end">
									<MenuItem>Edit</MenuItem>
									<MenuItem class="text-danger">Delete</MenuItem>
								</Menu>
							</Button>
						</Toggle>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<Pagination
		pagination={paginationStore}
		perPageOptions={[10, 25, 50]}
		show={["perPage", "pagination", "prevPage", "nextPage"]}
		classes={{
			root: "border-t py-1",
			perPage: "flex-1 text-right",
			pagination: "px-8",
		}}
	/>
</div>
