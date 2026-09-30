import { SvelteSet } from "svelte/reactivity";
import { Context, watch, type Getter } from "runed";
import { createMutation, createQueries, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { goto } from "$app/navigation";
import { resolve } from "$app/paths";
import { page } from "$app/state";
import { toast } from "svelte-sonner";
import {
	getIncidentOptions,
	type Incident,
	getSituationOptions,
	getInvestigationOptions,
	getInvestigationReportOptions,
	listSituationsQueryKey,
	requestSituationInvestigationMutation,
} from "$lib/api";
import {
	isDefinitiveUnavailableError,
	isProvisionalReport,
	reportSummary,
	SITUATION_POLL_INTERVAL_MS,
	type ReportState,
} from "$features/situations/lib/model";
import { situationHref } from "$features/situations/lib/routes";
import { investigationRunStatus } from "$features/situations/lib/status";
import { incidentResponseStatus } from "$features/incidents/lib/status";
import type { PageRelatedLink } from "$lib/app-shell.svelte";
import RiFireLine from "remixicon-svelte/icons/fire-line";

const idPath = (id?: string) => ({ id: id ?? "" });

export class SituationController {
	private queryClient = useQueryClient();
	situationId = $state<string>();

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.situationId = id;
		});
	}

	requestInvestigationMutation = createMutation(() => ({
		...requestSituationInvestigationMutation(),
		onSuccess: async (response, variables) => {
			const id = variables.path.id;
			this.queryClient.setQueryData(getSituationOptions({ path: { id } }).queryKey, response);
			await this.queryClient.invalidateQueries({ queryKey: listSituationsQueryKey() });
			if (this.situationId === id) {
				await goto(this.investigationHref);
			}
		},
		onError: (error) => {
			toast.error("Could not start investigation", { description: error.detail });
		},
	}));

	startInvestigation = () => {
		if (!this.situationId || this.investigationId || this.requestInvestigationMutation.isPending) {
			return;
		}
		this.requestInvestigationMutation.mutate({ path: { id: this.situationId } });
	};

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

	/** Loaded linked incidents in `linkedIncidentIds` order; loading or failed ones are omitted. */
	linkedIncidents = $derived(this.getLinkedIncidents());
	linkedIncidentsLoading = $derived(this.incidentsQuery.some((query) => query.isPending && !query.data));
	relatedLinks = $derived<PageRelatedLink[]>(
		this.linkedIncidents.map((incident) => ({
			key: incident.id,
			kind: "Incident",
			icon: RiFireLine,
			label: incident.attributes.title,
			path: resolve("/incidents/[slug]/[[view=incidentView]]", { slug: incident.attributes.slug }),
			status: incidentResponseStatus(incident.attributes.responseState),
		}))
	);

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

	runStatus = $derived(investigationRunStatus(this.investigationAttributes));
	reportAttributes = $derived(this.report?.attributes);
	reportState = $derived(this.getReportState());

	retryInvestigation = () => {
		this.investigationQuery.refetch();
	};

	retryReport = () => {
		this.reportQuery.refetch();
	};

	private getReportState(): ReportState {
		const report = this.reportAttributes;
		if (report) {
			return {
				kind: "published",
				attributes: report,
				summary: reportSummary(report),
				provisional: isProvisionalReport(report),
				refreshFailed: this.reportQuery.isError,
			};
		}
		if (this.reportAccessLost) {
			return { kind: "unavailable" };
		}
		if (this.reportQuery.isPending) {
			return { kind: "loading" };
		}
		if (this.reportQuery.isError && this.reportQuery.error?.status !== 404) {
			return { kind: "error" };
		}
		return { kind: "none" };
	}

	private getLinkedIncidents() {
		const incidents: Incident[] = [];
		for (const query of this.incidentsQuery) {
			const incident = query.data?.data;
			if (incident) {
				incidents.push(incident);
			}
		}
		return incidents;
	}
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
export const useSituationController = () => ctx.get();
