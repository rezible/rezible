<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Header, Pagination } from "svelte-ux";
	import { watch } from "runed";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { fromStore } from "svelte/store";
	import { listOncallEventsOptions, type OncallAnnotation, type OncallEvent, type OncallShift } from "$lib/api";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventsFilters, { type DisabledFilters, type FilterOptions } from "./EventsFilters.svelte";
	import EventRowItem from "./EventRowItem.svelte";

	type Props = {
		shift?: OncallShift;
		allowRosterActions?: string[];
		defaultFilters?: FilterOptions;
		disableFilters?: true | DisabledFilters;
	};
	const { shift, allowRosterActions, defaultFilters = {}, disableFilters = {} }: Props = $props();

	const defaultPerPage = 25;
	const paginationStore = createPaginationStore({ perPage: defaultPerPage });
	const pagination = fromStore(paginationStore);

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

	const annotationRosterId = $derived(shift?.attributes.roster.id ?? rosterIds?.at(0));
</script>

<div class="w-full h-full max-h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Events" subheading="subheading text" classes={{root: "p-2 w-full", title: "text-xl"}}>
		<svelte:fragment slot="actions">
			{#if disableFilters !== true}
				<div class="justify-end">
					<EventsFilters bind:filters disabled={disableFilters} />
				</div>
			{/if}
		</svelte:fragment>
	</Header>

	<div class="flex flex-col flex-1 overflow-y-auto">
		{#each pageData ?? [] as event}
			<EventRowItem 
				{event}
				annotationRosterIds={allowRosterActions}
				editAnnotation={(anno?: OncallAnnotation) => {setAnnotationDialog(event, anno)}}
			/>
		{/each}
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

{#if annotationRosterId}
	<EventAnnotationDialog 
		rosterId={annotationRosterId}
		event={annotationEvent} 
		current={annotationCurrent} 
		onClose={() => setAnnotationDialog()}
	/>
{/if}