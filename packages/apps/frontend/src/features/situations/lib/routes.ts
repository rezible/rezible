import { resolve } from "$app/paths";
import type { SituationViewParam } from "../../../params/situationView";

/** Links to a situation by its reference, or by its UUID, which its page replaces with the reference. */
export function situationHref(ref: string, view?: SituationViewParam, search = "") {
	return resolve("/situations/[ref]/[[view=situationView]]", { ref, view }) + search;
}

export function investigationHref(ref: string, search = "") {
	return situationHref(ref, "investigation", search);
}

/** Opens the situation with the declare incident dialog, prefilled with it. */
export function declareIncidentHref(ref: string) {
	return situationHref(ref, undefined, "?declare=incident");
}
