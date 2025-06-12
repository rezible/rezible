<script lang="ts">
	import { Field, ToggleGroup, ToggleOption } from "svelte-ux";
	import { createQuery, useQueryClient } from "@tanstack/svelte-query";
	import Header from "$components/header/Header.svelte";
	import { Button, DateRangeField, Pagination } from "svelte-ux";
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
	import { subMonths, subWeeks } from "date-fns";
	import { AnnotationDialogState, setAnnotationDialogState } from "$src/components/oncall-events/annotation-dialog/dialogState.svelte";

	type Props = {
		shift?: OncallShift;
		disableFilters?: true | DisabledFilters;
	};
	const { shift, disableFilters = {} }: Props = $props();

	const queryClient = useQueryClient();

	const oncallInfoQueryOpts = $derived(getUserOncallInformationOptions({query: { userId: session.userId }}));
	const oncallInfoQuery = createQuery(() => oncallInfoQueryOpts);
	const userOncallInfo = $derived(oncallInfoQuery.data?.data);
	const userRosterIds = $derived(userOncallInfo?.rosters.map(r => r.id) ?? []);
	const userActiveShifts = $derived(userOncallInfo?.activeShifts ?? []);
	const userActiveShift = $derived(userActiveShifts.at(0));

	const defaultShift = $derived(shift || userActiveShift);

	const today = new Date();
	const shiftDateRange = $derived(defaultShift && {
		from: new Date(defaultShift.attributes.startAt),
		to: new Date(defaultShift.attributes.endAt),
		periodType: PeriodType.Day,
	});
	const last7Days = {from: subWeeks(today, 1), to: today, periodType: PeriodType.Day};
	const lastMonth = {from: subMonths(today, 1), to: today, periodType: PeriodType.Day};

	type DateRangeOption = {label: string, value: "shift" | "7d" | "30d" | "custom"};
	const dateRangeOptions = [
		{label: "Last 7 Days", value: "7d"},
		{label: "Last Month", value: "30d"}, 
		{label: "Custom", value: "custom"},
	];
	let dateRangeOption = $state<DateRangeOption["value"]>("7d");
	watch(() => defaultShift, s => {dateRangeOption = (!!s) ? "shift" : dateRangeOption});

	let customDateRangeValue = $state<DateRangeType>(last7Days);
	const dateRange = $derived.by(() => {
		switch (dateRangeOption) {
			case "7d": return last7Days;
			case "30d": return lastMonth;
			case "shift": return (!!shiftDateRange ? shiftDateRange : customDateRangeValue);
			case "custom": return customDateRangeValue;
		}
	});

	let filters = $state<FilterOptions>({});
	let filtersVisible = $state(false);

	const defaultRosterId = $derived.by(() => {
		if (defaultShift) return defaultShift.attributes.roster.id;
		if (userRosterIds.length > 0) return userRosterIds.at(0);
	});
	watch(() => defaultRosterId, id => {filters.rosterId = id});

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

	const paginationStore = createPaginationStore({ page: 0, perPage: 25, total: 0 });
	const pagination = fromStore(paginationStore);
	const pageData = $derived(pagination.current.slice(eventsData));

	const numEvents = $derived(eventsData.length);
	watch(() => numEvents, paginationStore.setTotal);

	const annosQueryData = $derived<ListOncallAnnotationsData["query"]>({ 
		// TODO: backend
		// from: dateRange.from?.toISOString(),
		// to: dateRange.to?.toISOString(),
		rosterId: filters.rosterId,
	});
	const annosQueryOptions = $derived(listOncallAnnotationsOptions({query: annosQueryData}));
	const annosQuery = createQuery(() => ({...annosQueryOptions, enabled: !!userOncallInfo}));
	const annotations = $derived(annosQuery.data?.data ?? []);

	const eventAnnotations = $derived.by(() => {
		const annoMap = new Map<string, OncallAnnotation[]>();
		// TODO: ugly and probably slow
		annotations.forEach(ann => {
			const eventId = ann.attributes.event.id;
			const curr = annoMap.get(eventId) || [];
			annoMap.set(eventId, [...curr, ann]);
		});
		return annoMap;
	});

	const annoDialog = new AnnotationDialogState({
		onClosed: (updated?: OncallAnnotation) => {
			if (updated) queryClient.invalidateQueries(annosQueryOptions);
		}
	});
	setAnnotationDialogState(annoDialog);

	const loading = $derived(eventsQuery.isLoading || !userOncallInfo);
</script>

<EventAnnotationDialog />

<div class="w-full h-full max-h-full overflow-y-auto border rounded-lg flex flex-col">
	<Header title="Oncall Events" subheading="Operational events during these dates" classes={{root: "p-2 w-full", title: "text-xl"}}>
		{#snippet actions()}
			{#if disableFilters !== true}
				<div class="justify-end flex gap-2 items-end">
					<Field label="Date Range" labelPlacement="top" dense base classes={{root: "", container: "px-0 border-none py-0", input: "my-0 gap-2"}}>
						<ToggleGroup bind:value={dateRangeOption} variant="outline" inset classes={{root: "bg-surface-100"}}>
							{#if !!defaultShift}
								<ToggleOption value="shift">Active Shift</ToggleOption>
							{/if}

							{#each dateRangeOptions as opt}
								<ToggleOption value={opt.value}>{opt.label}</ToggleOption>
							{/each}
						</ToggleGroup>

						{#if dateRangeOption === "custom"}
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
								bind:value={() => dateRange, d => (customDateRangeValue = d)}
							/>
						{/if}
					</Field>

					<Button icon={mdiFilter} iconOnly 
						variant={filtersVisible ? "fill-light" : "default"}
						color={filtersVisible ? "accent" : "default"}
						on:click={() => {filtersVisible = !filtersVisible}} 
					/>
				</div>
			{/if}
		{/snippet}
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
				<EventRow {event} annotations={eventAnnotations.get(event.id)} />
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