import { useQueryClient, type QueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";

export const rosterIdCtx = new Context<string>("rosterId");

export class RosterViewState {
	rosterId = $state<string>();
}

const makeRosterState = () => {
	let rosterId = $state<string>();

	let queryClient = $state<QueryClient>();

	// const shiftQueryOpts = $derived(getOncallShiftOptions({ path: { id: (shiftId ?? "") } }))
	// const makeShiftQuery = () => createQuery(() => ({...shiftQueryOpts, enabled: !!shiftId}));
	// let shiftQuery = $state<ReturnType<typeof makeShiftQuery>>();

	// const shift = $derived(shiftQuery?.data?.data);


	const setup = (id: string) => {
		rosterId = id;
		queryClient = useQueryClient();
	}

	return {
		setup,
	}
}

export const rosterState = makeRosterState();