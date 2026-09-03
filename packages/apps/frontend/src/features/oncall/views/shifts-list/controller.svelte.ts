import { listOncallShiftsOptions } from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { subDays } from "date-fns";
import { Context } from "runed";
import { SvelteDate } from "svelte/reactivity";

const statusOptions = [
	{ label: "Active", value: "active" },
	{ label: "Past", value: "past" },
	{ label: "Upcoming", value: "upcoming", disabled: true },
];

export class OncallShiftsListController {
	selectedStatus = $state<string[]>(statusOptions.map((option) => option.value));
	dateRange = $state({
		from: new SvelteDate(subDays(Date.now(), 3).getTime()),
		to: new SvelteDate(),
		periodType: "day",
	});

	paginatedShiftsQuery = createPaginatedQuery({
		queryOptions: (pagination) => listOncallShiftsOptions({ query: pagination }),
		resetWhen: () => [$state.snapshot(this.selectedStatus), $state.snapshot(this.dateRange)],
	});

	query = $derived(this.paginatedShiftsQuery.query);

	onRosterSelected = (id?: string) => {
		if (!id) return;
	};

	setShiftStatus = (value?: string[]) => {
		if (!value?.length) return;
		this.selectedStatus = value;
	};
}

const ctx = new Context<OncallShiftsListController>("OncallShiftsListController");
export const initOncallShiftsListController = () => ctx.set(new OncallShiftsListController());
export const useOncallShiftsListController = () => ctx.get();
