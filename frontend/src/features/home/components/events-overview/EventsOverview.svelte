<script lang="ts">
	import { listOncallEventsOptions, type OncallShift } from "$lib/api";
	import { Button, Checkbox, Header, Icon, Menu, MenuItem, Pagination, Toggle, ToggleGroup, ToggleOption } from "svelte-ux";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { fromStore } from "svelte/store";
	import EventsFilters from "./EventsFilters.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import { mdiChevronDown, mdiChevronUp, mdiCircle, mdiCircleBoxOutline, mdiCircleOutline, mdiDotsVertical, mdiFilter } from "@mdi/js";
	import type { DateRange } from "@layerstack/utils/dateRange";
	import EventRow from "./EventRow.svelte";
	import { SvelteMap, SvelteSet } from "svelte/reactivity";

	type Props = {
		activeShift?: OncallShift;
	};
	const { activeShift }: Props = $props();

	const defaultPerPage = 10;
	const paginationStore = createPaginationStore({ perPage: defaultPerPage });
	const pagination = fromStore(paginationStore);

	type FilterOptions = {
		rosterIds?: string[];
		actionRequired?: boolean;
		dateRange?: DateRange;
	}
	let showCustomFilters = $state(false);
	let customFilters = $state<FilterOptions>({});

	// TODO
	const queryFilters = $derived(customFilters);

	const eventsQuery = createQuery(() => listOncallEventsOptions({ query: { 
		rosterIds: customFilters.rosterIds,
	}}));
	const eventsData = $derived(eventsQuery.data?.data ?? []);

	const totalEvents = $derived(eventsData.length);
	$effect(() => paginationStore.setTotal(totalEvents));

	const pageData = $derived(pagination.current.slice(eventsData));

	const checked = $derived(new SvelteSet<string>());
	const eventCheckToggle = (id: string) => {
		if (checked.has(id)) checked.delete(id);
		else checked.add(id);	
	}
	const anyChecked = $derived(checked.size > 0);
	const onMasterCheckToggle = () => {
		if (anyChecked) {
			checked.clear();
		} else {
			pageData.forEach(e => checked.add(e.id));
		}
	}
</script>

<div class="w-full h-full max-h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Events" classes={{root: "p-2", title: "text-xl"}}>
		<div slot="actions" class="flex items-center gap-2">
			<Button variant={"default"} on:click={() => (showCustomFilters = !showCustomFilters)}>
				<Icon data={mdiFilter} />
				Filter
				<Icon data={showCustomFilters ? mdiChevronUp : mdiChevronDown} />
			</Button>
		</div>
	</Header>

	{#if showCustomFilters}
		<div class="p-2 pt-0 w-full">
			<EventsFilters 
				bind:rosterIds={customFilters.rosterIds}
				bind:actionRequired={customFilters.actionRequired}
				bind:dateRange={customFilters.dateRange}
			/>
		</div>
	{/if}

	<div class="flex flex-col overflow-y-auto">
		<div class="grid grid-flow-row grid-cols-[48px_64px_minmax(100px,1fr)_64px] min-h-0 overflow-y-auto">
			<div class="grid grid-cols-subgrid col-span-full sticky top-0 bg-surface-100 items-center p-2">
				<div class="grid place-self-center"><Checkbox checked={anyChecked} indeterminate on:change={onMasterCheckToggle} /></div>
				<span class="grid place-self-center">Kind</span>
				<span class="pl-2">Title</span>
				<span class="">Actions</span>
			</div>

			{#each pageData ?? [] as event}
				<EventRow {event} checked={checked.has(event.id)} onToggleChecked={() => {eventCheckToggle(event.id)}} />
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
