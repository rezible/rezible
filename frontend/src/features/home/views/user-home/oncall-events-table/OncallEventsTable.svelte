<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Header, Pagination } from "svelte-ux";
	import { watch } from "runed";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { fromStore } from "svelte/store";
	import { session } from "$lib/auth.svelte";
	import { getUserOncallInformationOptions, listOncallEventsOptions, type ListOncallEventsData, type OncallAnnotation, type OncallEvent, type OncallShift } from "$lib/api";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventsFilters, { type DisabledFilters, type FilterOptions } from "$components/oncall-events/EventsFilters.svelte";
	import EventRowItem from "$components/oncall-events/EventRowItem.svelte";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import { parseAbsoluteToLocal } from "@internationalized/date";

	type Props = {
		shift?: OncallShift;
		disableFilters?: true | DisabledFilters;
	};
	const { shift, disableFilters = {} }: Props = $props();

	const oncallInfoQueryOpts = $derived(getUserOncallInformationOptions({query: { userId: session.userId }}));
	const oncallInfoQuery = createQuery(() => oncallInfoQueryOpts);
	const userOncallInfo = $derived(oncallInfoQuery.data?.data);

	const userRosterIds = $derived(userOncallInfo?.rosters.map(r => r.id) ?? []);
	const userActiveShifts = $derived(userOncallInfo?.activeShifts ?? []);
	const userActiveShift = $derived(userActiveShifts.at(0));

	const ONE_DAY = (1000 * 60 * 60 * 24);

	const defaultShift = $derived(shift || userActiveShift);
	const defaultFilters = $derived.by(() => {
		let f: FilterOptions = {};
		if (defaultShift) {
			f.rosterId = defaultShift.attributes.roster.id;
			f.dateRange = {
				from: new Date(defaultShift.attributes.startAt),
				to: new Date(defaultShift.attributes.endAt),
			}
		} else if (userRosterIds.length > 0) {
			f.rosterId = userRosterIds.at(0);
			f.dateRange = {
				from: new Date(Date.now() - ONE_DAY),
				to: new Date(),
			}
		}
		return f;
	});

	let filters = $state<FilterOptions>({});
	watch(() => defaultFilters, f => {filters = f});

	const paginationStore = createPaginationStore({ perPage: 25 });
	const pagination = fromStore(paginationStore);

	const eventsQueryData = $derived<ListOncallEventsData["query"]>({ 
		from: filters.dateRange?.from?.toISOString(),
		to: filters.dateRange?.to?.toISOString(),
		rosterId: filters.rosterId,
		withAnnotations: true,
	})
	const eventsQuery = createQuery(() => ({
		...listOncallEventsOptions({query: eventsQueryData}),
		enabled: !!userOncallInfo,
	}));
	const eventsData = $derived(eventsQuery.data?.data ?? []);

	watch(() => eventsData.length, num => {paginationStore.setTotal(num)});
	const pageData = $derived(pagination.current.slice(eventsData));

	let annotationEvent = $state<OncallEvent>();
	let annotationCurrent = $state<OncallAnnotation>();
	const setAnnotationDialog = (ev?: OncallEvent, anno?: OncallAnnotation) => {
		annotationEvent = ev;
		annotationCurrent = anno;
	};

	const annotationRoster = $derived(defaultShift?.attributes.roster);

	const loading = $derived(eventsQuery.isLoading || !userOncallInfo);
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
		{#if loading}
			<LoadingIndicator />
		{:else}
			{#each pageData ?? [] as event}
				<EventRowItem 
					{event}
					annotationRosterIds={userRosterIds}
					editAnnotation={(anno?: OncallAnnotation) => {setAnnotationDialog(event, anno)}}
				/>
			{/each}
		{/if}
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

{#if annotationRoster}
	<EventAnnotationDialog 
		roster={annotationRoster}
		event={annotationEvent} 
		current={annotationCurrent} 
		onClose={() => setAnnotationDialog()}
	/>
{/if}