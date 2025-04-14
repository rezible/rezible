<script lang="ts">
	import { listOncallEventsOptions, type OncallEvent } from "$lib/api";
	import { Checkbox, Header, Pagination } from "svelte-ux";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { fromStore } from "svelte/store";
	import EventsFilters, { type FilterOptions } from "./EventsFilters.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import EventRow from "./EventRow.svelte";
	import { SvelteSet } from "svelte/reactivity";
	import AnnotationDialog from "./AnnotationDialog.svelte";

	type Props = {
	};
	const { }: Props = $props();

	const defaultPerPage = 25;
	const paginationStore = createPaginationStore({ perPage: defaultPerPage });
	const pagination = fromStore(paginationStore);

	let filters = $state<FilterOptions>({});

	const rosterIds = $derived((filters.rosterIds && filters.rosterIds.length > 0) ? filters.rosterIds : undefined);
	const eventsQuery = createQuery(() => listOncallEventsOptions({ query: { rosterIds }}));
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

	let annotateEvent = $state<OncallEvent>();
	const onOpenAnnotateDialog = (event: OncallEvent) => {
		annotateEvent = event;
	}
</script>

<div class="w-full h-full max-h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Events" subheading="subheading text" classes={{root: "p-2 w-full", title: "text-xl"}}>
		<div slot="actions" class="p-2 pt-0">
			<EventsFilters 
				bind:filters
			/>
		</div>
	</Header>

	<div class="flex flex-col overflow-y-auto">
		<div class="grid grid-flow-row grid-cols-[48px_64px_140px_minmax(200px,auto)_minmax(100px,1fr)_minmax(100px,auto)] gap-x-2 min-h-0 overflow-y-auto">
			<div class="grid grid-cols-subgrid col-span-full sticky top-0 bg-surface-100 items-center p-2">
				<div class="grid place-self-center"><Checkbox checked={anyChecked} indeterminate on:change={onMasterCheckToggle} /></div>
				<span class="grid place-items-center">Kind</span>
				<span class="flex items-center">Time</span>
				<span class="flex items-center">Title</span>
				<span class="flex items-center">Annotation</span>
				<span class="flex items-center justify-end">Actions</span>
			</div>

			{#each pageData ?? [] as event}
				<EventRow 
					{event} 
					checked={checked.has(event.id)} 
					onToggleChecked={() => {eventCheckToggle(event.id)}}
					onOpenAnnotateDialog={() => {onOpenAnnotateDialog(event)}}
				/>
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

<AnnotationDialog event={annotateEvent} onClose={() => (annotateEvent = undefined)} />