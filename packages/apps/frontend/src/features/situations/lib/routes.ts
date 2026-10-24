import { resolve } from "$app/paths";
import type { SituationViewParam } from "../../../params/situationView";

export function situationHref(id: string, view?: SituationViewParam, search = "") {
	return resolve("/situations/[id]/[[view=situationView]]", { id, view }) + search;
}

export function investigationHref(situationId: string, search = "") {
	return situationHref(situationId, "investigation", search);
}

/** Opens the situation with the declare incident dialog, prefilled with it. */
export function declareIncidentHref(situationId: string) {
	return situationHref(situationId, undefined, "?declare=incident");
}
