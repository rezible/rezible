import { SvelteSet } from "svelte/reactivity";
import { createQueries, createQuery } from "@tanstack/svelte-query";
import { page } from "$app/state";
import { replaceState } from "$app/navigation";
import {
	getIncidentOptions,
	getSituationOptions,
	getSituationInvestigationOptions,
	listSituationInvestigationsOptions,
} from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import { Context, watch, type Getter } from "runed";
import { investigationSearch, selectInvestigation } from "./model";
import { situationHref, investigationHref } from "../../lib/routes";

const idPath = (id?: string) => ({ id: id ?? "" });

export class SituationController {
	situationId = $state<string>();

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.situationId = id;
		});
		watch(
			() => [this.investigationViewActive, this.selectedInvestigationId, page.url.search] as const,
			([active, selectedId]) => {
				if (!active || !selectedId || page.url.searchParams.get("investigation")) {
					return;
				}
				const search = investigationSearch(page.url.search, selectedId);
				replaceState(`${page.url.pathname}${search}${page.url.hash}`, page.state);
			}
		);
	}

	situationQuery = createQuery(() => ({
		...getSituationOptions({ path: idPath(this.situationId) }),
		enabled: !!this.situationId,
	}));

	situation = $derived(this.situationQuery.data?.data);
	incidentIds = $derived([...new SvelteSet(this.situation?.attributes?.linkedIncidentIds ?? [])]);
	incidentsQuery = createQueries(() => ({
		queries: this.incidentIds.map((id) => getIncidentOptions({ path: { id } })),
	}));

	private investigationViewActive = $derived(
		page.params.view === "analysis" || page.params.view === "investigations"
	);

	selectedInvestigationId = $derived(
		selectInvestigation(
			page.url.searchParams.get("investigation"),
			undefined,
			this.situation?.attributes.investigations ?? []
		)
	);

	investigationsQuery = createPaginatedQuery({
		resetWhen: () => this.situationId,
		keepPreviousQueryData: false,
		queryOptions: (pagination) => ({
			...listSituationInvestigationsOptions({ path: idPath(this.situationId), query: pagination }),
			enabled: !!this.situation && this.investigationViewActive,
		}),
	});
	investigations = $derived(this.investigationsQuery.query.data?.data ?? []);

	selectedInvestigationQuery = createQuery(() => ({
		...getSituationInvestigationOptions({ path: idPath(this.selectedInvestigationId) }),
		enabled: !!this.selectedInvestigationId && this.investigationViewActive,
	}));
	selectedInvestigation = $derived(this.selectedInvestigationQuery.data?.data);
	selectedInvestigationBelongs = $derived(
		!!this.selectedInvestigation && this.selectedInvestigation.attributes.situationId === this.situationId
	);

	investigationAnalysisId = $derived(
		this.selectedInvestigationBelongs ? this.selectedInvestigation?.attributes.analysisId : undefined
	);

	private id = $derived(this.situationId ?? "");
	investigationsHref = $derived(situationHref(this.id, "investigations", page.url.search));

	investigationSelectionHref(id: string) {
		const param = page.params.view === "impact" ? "impact" : "investigations";
		return situationHref(this.id, param, investigationSearch(page.url.search, id));
	}
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
export const useSituationController = () => ctx.get();
