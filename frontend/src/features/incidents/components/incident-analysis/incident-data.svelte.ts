import { listIncidentMilestonesOptions, listIncidentSystemComponentsOptions, type Incident, type ListIncidentMilestonesResponse, type ListIncidentSystemComponentsResponse } from "$src/lib/api";
import { createQuery, type CreateQueryResult, QueryClient, useQueryClient } from "@tanstack/svelte-query";
import { watch } from "runed";

const createIncidentDataState = () => {
	let incidentId = $state("");
	const enabled = $derived(!!incidentId);

	const setIncidentId = (id: string) => {
		incidentId = id;
	}

	const milestonesQueryOpts = $derived(({
		...listIncidentMilestonesOptions({path: {id: incidentId}}),
		enabled,
	}));
	let milestonesQuery = $state<ReturnType<typeof createQuery<ListIncidentMilestonesResponse>>>();
	const milestones = $derived(milestonesQuery?.data?.data ?? []);

	const incidentComponentsQueryOpts = $derived(({
		...listIncidentSystemComponentsOptions({path: {id: incidentId}}),
		enabled,
	}));
	let incidentComponentsQuery = $state<ReturnType<typeof createQuery<ListIncidentSystemComponentsResponse>>>();
	const incidentComponents = $derived(incidentComponentsQuery?.data?.data ?? []);

	let eventsQuery = $state<CreateQueryResult<ListIncidentMilestonesResponse, Error>>();
	const events = $derived(eventsQuery?.data?.data ?? []);

	const mount = (incidentIdFn: () => string) => {
		const queryClient = useQueryClient();

		milestonesQuery = createQuery(() => milestonesQueryOpts, queryClient);
		incidentComponentsQuery = createQuery(() => incidentComponentsQueryOpts, queryClient);

		watch(incidentIdFn, setIncidentId);
	}

	return {
		mount,
		get milestones() {return milestones},
		get incidentComponents() {return incidentComponents},
		get events() {return events},
	}
}
export const incidentData = createIncidentDataState();