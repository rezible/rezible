import type { ListEventsData, EventAttributes } from "$lib/api";
import { useUserOncallInformation } from "$lib/userOncall.svelte";
import { subMonths, subWeeks } from "date-fns";

export type DateRangeOption = { label: string, value: "shift" | "7d" | "30d" | "custom" };

const last7Days = () => ({ from: subWeeks(new Date(), 1), to: new Date(), periodType: "day" });
const lastMonth = () => ({ from: subMonths(new Date(), 1), to: new Date(), periodType: "day" });

export type EventKind = EventAttributes["kind"];

export type FilterOptions = {
	rosterId?: string;
	eventKinds?: EventKind[];
	annotated?: boolean;
};

export class EventsListFiltersState {
	private oncallInfo = useUserOncallInformation();
	activeShift = $derived(this.oncallInfo.activeShifts.at(0));
	private defaultShiftDateRange = $derived(this.activeShift && {
		from: new Date(this.activeShift.attributes.startAt),
		to: new Date(this.activeShift.attributes.endAt),
		periodType: "day",
	});

	dateRangeOption = $state<DateRangeOption["value"]>("7d");
	customDateRangeValue = $state(last7Days());
	shiftDateRange = $derived(!!this.defaultShiftDateRange ? this.defaultShiftDateRange : this.customDateRangeValue)

	dateRange = $derived.by(() => {
		switch (this.dateRangeOption) {
			case "7d": return last7Days();
			case "30d": return lastMonth();
			case "shift": return this.shiftDateRange;
			case "custom": return this.customDateRangeValue;
		}
	});
	
	private defaultRosterId = $derived.by(() => {
		if (this.activeShift) return this.activeShift.attributes.roster.id;
		if (this.oncallInfo.rosterIds.length > 0) return this.oncallInfo.rosterIds.at(0);
	});

	eventKinds = $state<EventKind[]>();
	annotation = $state<boolean>();
	rosterId = $state<string>();

	private listRosterEventsQueryData = $derived<ListEventsData["query"]>({
		from: this.dateRange.from?.toISOString(),
		to: this.dateRange.to?.toISOString(),
		// rosterId: this.rosterId,
	});
	private listShiftEventsQueryData = $derived<ListEventsData["query"]>({ 
		// shiftId: this.activeShift?.id 
	});
	queryData = $derived(this.dateRangeOption === "shift" ? this.listShiftEventsQueryData : this.listRosterEventsQueryData);
	queryEnabled = $derived(!!this.oncallInfo && !!this.defaultRosterId);
};