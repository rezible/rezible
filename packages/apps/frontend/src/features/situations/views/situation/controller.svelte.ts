import { SvelteSet } from "svelte/reactivity";
import { Context, watch, type Getter } from "runed";
import { createQueries, createQuery } from "@tanstack/svelte-query";
import { page } from "$app/state";
import {
	getIncidentOptions,
	getSituationOptions,
	getInvestigationOptions,
	getInvestigationReportOptions,
} from "$lib/api";
import {
	isDefinitiveUnavailableError,
	investigationExecution,
	timestamp,
	SITUATION_POLL_INTERVAL_MS,
} from "$features/situations/lib/model";
import { situationHref } from "$features/situations/lib/routes";

const idPath = (id?: string) => ({ id: id ?? "" });

export class SituationController {
	situationId = $state<string>();

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.situationId = id;
		});
	}

	situationQuery = createQuery(() => ({
		...getSituationOptions({ path: idPath(this.situationId) }),
		enabled: !!this.situationId,
		refetchInterval: SITUATION_POLL_INTERVAL_MS,
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	situation = $derived(
		isDefinitiveUnavailableError(this.situationQuery.error) ? undefined : this.situationQuery.data?.data
	);
	situationUnavailable = $derived(isDefinitiveUnavailableError(this.situationQuery.error));
	incidentIds = $derived([...new SvelteSet(this.situation?.attributes.linkedIncidentIds ?? [])]);
	incidentsQuery = createQueries(() => ({
		queries: this.incidentIds.map((id) => getIncidentOptions({ path: { id } })),
	}));

	situationInvestigation = $derived(this.situation?.attributes.investigation);
	investigationId = $derived(this.situationInvestigation?.investigation.id);
	investigationHref = $derived(situationHref(this.situationId ?? "", "investigation", page.url.search));

	investigationQuery = createQuery(() => ({
		...getInvestigationOptions({ path: idPath(this.investigationId) }),
		enabled: !!this.investigationId,
		refetchInterval: (query) => {
			return isDefinitiveUnavailableError(query.state.error) ? false : SITUATION_POLL_INTERVAL_MS;
		},
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	investigation = $derived(
		isDefinitiveUnavailableError(this.investigationQuery.error)
			? undefined
			: this.investigationQuery.data?.data
	);
	investigationUnavailable = $derived(isDefinitiveUnavailableError(this.investigationQuery.error));
	investigationAttributes = $derived(this.investigation?.attributes);

	reportQuery = createQuery(() => ({
		...getInvestigationReportOptions({ path: idPath(this.investigationId) }),
		enabled: !!this.investigationId,
		refetchInterval: (query) => {
			if ([401, 403].includes(query.state.error?.status ?? 0) || this.investigationUnavailable) {
				return false;
			}
			return SITUATION_POLL_INTERVAL_MS;
		},
		refetchIntervalInBackground: false,
		retry: (failureCount, error) => !isDefinitiveUnavailableError(error) && failureCount < 2,
	}));

	report = $derived(
		this.investigationUnavailable || isDefinitiveUnavailableError(this.reportQuery.error)
			? undefined
			: this.reportQuery.data?.data
	);
	reportAccessLost = $derived([401, 403].includes(this.reportQuery.error?.status ?? 0));

	execution = $derived(investigationExecution(this.investigationAttributes));
	reportAttributes = $derived(this.report?.attributes);
	reportPublishedAt = $derived(timestamp(this.reportAttributes?.createdAt));
	reportReferences = $derived(this.reportAttributes?.references ?? []);
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
export const useSituationController = () => ctx.get();
