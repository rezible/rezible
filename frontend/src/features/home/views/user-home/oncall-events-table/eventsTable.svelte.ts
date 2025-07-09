import { listOncallAnnotationsOptions, listOncallEventsOptions, type ListOncallAnnotationsData, type ListOncallEventsData, type OncallAnnotation } from "$lib/api";
import type { FilterOptions } from "$src/components/oncall-events/EventsFilters.svelte";
import { PeriodType } from "@layerstack/utils";
import { paginationStore as createPaginationStore } from "@layerstack/svelte-stores";
import { type DateRange as DateRangeType } from "@layerstack/utils/dateRange";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { subMonths, subWeeks } from "date-fns";
import { watch } from "runed";
import { fromStore } from "svelte/store";
import { useUserOncallInformation } from "$lib/userOncall.svelte";

export type DateRangeOption = {label: string, value: "shift" | "7d" | "30d" | "custom"};

const last7Days = () => ({from: subWeeks(new Date(), 1), to: new Date(), periodType: PeriodType.Day});
const lastMonth = () => ({from: subMonths(new Date(), 1), to: new Date(), periodType: PeriodType.Day});

export class OncallEventsTableState {
	private queryClient = useQueryClient();
	private oncallInfo = useUserOncallInformation();

	activeShift = $derived(this.oncallInfo.activeShifts.at(0));

	defaultShiftDateRange = $derived(this.activeShift && {
		from: new Date(this.activeShift.attributes.startAt),
		to: new Date(this.activeShift.attributes.endAt),
		periodType: PeriodType.Day,
	});

	dateRangeOption = $state<DateRangeOption["value"]>("7d");
	customDateRangeValue = $state<DateRangeType>(last7Days());
	shiftDateRange = $derived(!!this.defaultShiftDateRange ? this.defaultShiftDateRange : this.customDateRangeValue)

	dateRange = $derived.by(() => {
		switch (this.dateRangeOption) {
			case "7d": return last7Days();
			case "30d": return lastMonth();
			case "shift": return this.shiftDateRange;
			case "custom": return this.customDateRangeValue;
		}
	});

	filters = $state<FilterOptions>({});
	defaultRosterId = $derived.by(() => {
		if (this.activeShift) return this.activeShift.attributes.roster.id;
		if (this.oncallInfo.rosterIds.length > 0) return this.oncallInfo.rosterIds.at(0);
	});

	paginationStore = createPaginationStore({ page: 1, perPage: 10, total: 0 });
	pagination = fromStore(this.paginationStore);

	paginationTotal = $derived(this.pagination.current.total);
	paginationPerPage = $derived(this.pagination.current.perPage);
	paginationCurrentPage = $derived(this.pagination.current.page as number);

	queryEnabled = $derived(!!this.oncallInfo && !!this.defaultRosterId);
	
	private listEventsQueryOffset = $derived(Math.max(0, (this.paginationCurrentPage - 1) * this.paginationPerPage));
	private isShiftQuery = $derived(this.dateRangeOption === "shift");
	private listEventsQueryShiftId = $derived((this.isShiftQuery && this.activeShift) ? this.activeShift.id : undefined);
	private listEventsQueryData = $derived<ListOncallEventsData["query"]>({ 
		from: this.isShiftQuery ? undefined : this.dateRange.from?.toISOString(),
		to: this.isShiftQuery ? undefined : this.dateRange.to?.toISOString(),
		rosterId: this.isShiftQuery ? undefined : this.filters.rosterId,
		shiftId: this.listEventsQueryShiftId,
		limit: this.paginationPerPage,
		offset: this.listEventsQueryOffset,
		withAnnotations: true,
	});
	private listEventsQueryOptions = $derived(listOncallEventsOptions({query: this.listEventsQueryData}));
	private listEventsQuery = createQuery(() => ({
		...this.listEventsQueryOptions,
		enabled: this.queryEnabled,
	}));
	private eventsQueryData = $derived(this.listEventsQuery.data);
	private eventsData = $derived(this.eventsQueryData?.data ?? []);

	pageData = $derived(this.eventsData);

	invalidateQuery() {
		this.queryClient.invalidateQueries(this.listEventsQueryOptions);
	}

	loading = $derived(this.listEventsQuery.isLoading || !this.oncallInfo.loaded)

	constructor() {
		watch(() => this.activeShift, s => {
			this.dateRangeOption = !!s ? "shift" : this.dateRangeOption;
		});

		watch(() => this.defaultRosterId, id => {
			this.filters.rosterId = id;
		});

		watch(() => this.eventsQueryData, d => {
			if (!d) return;
			this.paginationStore.setTotal(d.pagination.total ?? 0);
		});
	};
}
