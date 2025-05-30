<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { Button, DateRangeField, Header, Pagination } from "svelte-ux";
	import { watch } from "runed";
	import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
	import { fromStore } from "svelte/store";
	import { session } from "$lib/auth.svelte";
	import { getUserOncallInformationOptions, listOncallAnnotationsOptions, listOncallEventsOptions, type ListOncallAnnotationsData, type ListOncallEventsData, type OncallAnnotation, type OncallEvent, type OncallShift } from "$lib/api";
	import EventAnnotationDialog from "$components/oncall-events/annotation-dialog/EventAnnotationDialog.svelte";
	import EventsFilters, { type DisabledFilters, type FilterOptions } from "$components/oncall-events/EventsFilters.svelte";
	import EventRow from "$components/oncall-events/EventRow.svelte";
	import LoadingIndicator from "$components/loader/LoadingIndicator.svelte";
	import { type DateRange as DateRangeType } from "@layerstack/utils/dateRange";
	import { PeriodType } from "@layerstack/utils";
	import { mdiCalendarRange, mdiFilter } from "@mdi/js";
	import { subDays } from "date-fns";

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

	const today = new Date();
	const last7Days = {from: subDays(today, 7), to: today, periodType: PeriodType.Day};

	const defaultShift = $derived(shift || userActiveShift);

	const ONE_DAY = (1000 * 60 * 60 * 24);
	const defaultDateRange = $derived.by<DateRangeType>(() => {
		if (defaultShift) {
			return {
				from: new Date(defaultShift.attributes.startAt),
				to: new Date(defaultShift.attributes.endAt),
				periodType: PeriodType.Day,
			}
		} else if (userRosterIds.length > 0) {
			return {
				from: new Date(Date.now() - ONE_DAY),
				to: new Date(),
				periodType: PeriodType.Day,
			}
		}
		return last7Days;
	});

	let dateRangeValue = $state<DateRangeType>();
	const dateRange = $derived(dateRangeValue || defaultDateRange);

	const defaultRosterId = $derived.by(() => {
		if (defaultShift) return defaultShift.attributes.roster.id;
		if (userRosterIds.length > 0) return userRosterIds.at(0);
	});
	let filters = $state<FilterOptions>({});
	watch(() => defaultRosterId, id => {filters.rosterId = id});
	let filtersVisible = $state(false);

	const paginationStore = createPaginationStore({ page: 0, perPage: 25, total: 0 });
	const pagination = fromStore(paginationStore);

	const eventsQueryData = $derived<ListOncallEventsData["query"]>({ 
		from: dateRange.from?.toISOString(),
		to: dateRange.to?.toISOString(),
		rosterId: filters.rosterId,
	});
	const eventsQuery = createQuery(() => ({
		...listOncallEventsOptions({query: eventsQueryData}),
		enabled: !!userOncallInfo,
	}));
	const eventsData = $derived(eventsQuery.data?.data ?? []);
	const numEvents = $derived(eventsData.length);
	watch(() => numEvents, paginationStore.setTotal);
	const pageData = $derived(pagination.current.slice(eventsData));

	const annosQueryData = $derived<ListOncallAnnotationsData["query"]>({ 
		// TODO: backend
		// from: dateRange.from?.toISOString(),
		// to: dateRange.to?.toISOString(),
		rosterId: filters.rosterId,
	});
	const annosQuery = createQuery(() => ({
		...listOncallAnnotationsOptions({query: annosQueryData}),
		enabled: !!userOncallInfo,
	}));
	const annosData = $derived(annosQuery.data?.data ?? []);
	const eventAnnotations = $derived.by(() => {
		const annoMap = new Map<string, OncallAnnotation[]>();
		// TODO: ugly and probably slow
		annosData.forEach(ann => {
			const eventId = ann.attributes.event.id;
			annoMap.set(eventId, [...(annoMap.get(eventId) || []), ann]);
		});
		return annoMap;
	});

	let annoDialogEvent = $state<OncallEvent>();
	let annoDialogAnno = $state<OncallAnnotation>();
	const setAnnotationDialog = (ev?: OncallEvent, anno?: OncallAnnotation) => {
		annoDialogEvent = ev;
		annoDialogAnno = anno;
	};

	const annotationRoster = $derived(defaultShift?.attributes.roster);

	const loading = $derived(eventsQuery.isLoading || !userOncallInfo);
</script>

<div class="w-full h-full max-h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Oncall Events" subheading="Operational events during these dates" classes={{root: "p-2 w-full", title: "text-xl"}}>
		<svelte:fragment slot="actions">
			{#if disableFilters !== true}
				<div class="justify-end flex gap-2 items-end">
					<DateRangeField
						label="Date Range"
						labelPlacement="top"
						periodTypes={[PeriodType.Day]}
						dense
						classes={{
							field: { root: "gap-0", container: "pl-0 h-8 flex items-center", prepend: "[&>span]:mr-2" },
						}}
						icon={mdiCalendarRange}
						bind:value={() => dateRange, d => (dateRangeValue = d)}
					/>

					<Button icon={mdiFilter} iconOnly 
						variant={filtersVisible ? "fill-light" : "default"}
						color={filtersVisible ? "accent" : "default"}
						on:click={() => {filtersVisible = !filtersVisible}} 
					/>
				</div>
			{/if}
		</svelte:fragment>
	</Header>

	{#if disableFilters !== true && filtersVisible}
		<div class="w-full p-2 pt-0">
			<EventsFilters bind:filters disabled={disableFilters} />
		</div>
	{/if}

	<div class="flex flex-col flex-1 overflow-y-auto border-t">
		{#if loading}
			<LoadingIndicator />
		{:else}
			{#each pageData as event}
				<EventRow 
					{event}
					annotations={eventAnnotations.get(event.id)}
					annotatableRosterIds={userRosterIds}
					editAnnotation={anno => {setAnnotationDialog(event, anno)}}
				/>
			{:else}
				<div class="grid place-items-center flex-1">
					<span class="text-surface-content/80">No Events</span>
				</div>
			{/each}
		{/if}
	</div>

	{#if numEvents > 0}
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
	{/if}
</div>

{#if annotationRoster}
	<EventAnnotationDialog 
		roster={annotationRoster}
		event={annoDialogEvent} 
		current={annoDialogAnno} 
		onClose={() => setAnnotationDialog()}
	/>
{/if}