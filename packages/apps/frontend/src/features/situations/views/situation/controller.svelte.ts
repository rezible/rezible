import { SvelteSet } from "svelte/reactivity";
import { createQueries, createQuery } from "@tanstack/svelte-query";
import { getIncidentOptions, getSituationOptions } from "$lib/api";
import { Context, watch, type Getter } from "runed";

export class SituationController {
	id = $state<string>(null!);
	query = createQuery(() => getSituationOptions({ path: { id: this.id } }));

	constructor(idFn: Getter<string>) {
		this.id = idFn();
		watch(idFn, id => {
			this.id = id;
		})
	}

	situation = $derived(this.query.data?.data);
	incidentIds = $derived([...new SvelteSet(this.situation?.attributes?.linkedIncidentIds ?? [])]);
	incidentsQuery = createQueries(() => ({
		queries: this.incidentIds.map(id => getIncidentOptions({ path: { id } })),
	}));
}

const ctx = new Context<SituationController>("SituationController");
export const initSituationController = (idFn: Getter<string>) => ctx.set(new SituationController(idFn));
export const useSituationController = () => ctx.get();
