<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, Header, Icon, Pagination } from "svelte-ux";
	import { watch } from "runed";
	import { mdiFilter } from "@mdi/js";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { fromStore } from "svelte/store";
	import { listOncallEventsOptions, type OncallAnnotation, type OncallEvent, type OncallShift } from "$lib/api";
	import EventAnnotationDialog from "$components/event-annotation-dialog/EventAnnotationDialog.svelte";
	import EventsFilters, { type DisabledFilters, type FilterOptions } from "./EventsFilters.svelte";
	import EventRow from "./EventRow.svelte";

	type Props = {
		shift?: OncallShift;
		defaultFilters?: FilterOptions;
		allowRosterActions?: string[];
		disableFilters?: true | DisabledFilters;
	};
	const { shift, defaultFilters = {}, allowRosterActions = [], disableFilters = {} }: Props = $props();

	const defaultPerPage = 25;
	const paginationStore = createPaginationStore({ perPage: defaultPerPage });
	const pagination = fromStore(paginationStore);

	let showFilters = $state(false);

	let filters = $state<FilterOptions>(defaultFilters);
	watch(() => defaultFilters, (f) => {filters = f});

	const rosterIds = $derived((filters.rosterIds && filters.rosterIds.length > 0) ? filters.rosterIds : undefined);
	const eventsQuery = createQuery(() => listOncallEventsOptions({ query: { rosterIds }}));
	const eventsData = $derived(eventsQuery.data?.data ?? []);

	const totalEvents = $derived(eventsData.length);
	$effect(() => paginationStore.setTotal(totalEvents));

	const pageData = $derived(pagination.current.slice(eventsData));

	let annotationEvent = $state<OncallEvent>();
	let annotationCurrent = $state<OncallAnnotation>();
	const setAnnotationDialog = (ev?: OncallEvent, anno?: OncallAnnotation) => {
		annotationEvent = ev;
		annotationCurrent = anno;
	};
</script>

<div class="w-full h-full max-h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Events" subheading="subheading text" classes={{root: "p-2 w-full", title: "text-xl"}}>
		<svelte:fragment slot="actions">
			{#if disableFilters !== true}
				<Button color={showFilters ? "accent" : "default"} variant={showFilters ? "fill-light" : "default"} on:click={() => (showFilters = !showFilters)}>
					<span class="flex gap-2 items-center">Filters <Icon data={mdiFilter} /></span>
				</Button>
			{/if}
		</svelte:fragment>
	</Header>

	{#if disableFilters !== true}
		<div class="p-2 border-t justify-end" class:hidden={!showFilters}>
			<EventsFilters bind:filters disabled={disableFilters} />
		</div>
	{/if}

	<div class="flex flex-col flex-1 overflow-y-auto">
		<div class="grid grid-flow-row grid-cols-[64px_128px_minmax(150px,auto)_minmax(100px,1fr)] gap-x-2 min-h-0 overflow-y-auto">
			<div class="grid grid-cols-subgrid col-span-full sticky top-0 bg-surface-100 items-center p-2">
				<span class="grid place-items-center">Kind</span>
				<span class="flex items-center">Time</span>
				<span class="flex items-center">Title</span>
				<span class="flex items-center justify-end">Annotations</span>
			</div>

			{#each pageData ?? [] as event}
				<EventRow 
					{event}
					allowAnnotationRosters={allowRosterActions}
					onOpenAnnotateDialog={(anno?: OncallAnnotation) => {setAnnotationDialog(event, anno)}}
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

<EventAnnotationDialog event={annotationEvent} current={annotationCurrent} onClose={() => setAnnotationDialog()} />