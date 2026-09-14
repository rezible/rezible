import { resolve } from "$app/paths";
import type { SituationViewParam } from "../../../params/situationView";
import { investigationSearch } from "../views/situation/model";

export function situationHref(id: string, view?: SituationViewParam, search = "") {
	return resolve("/situations/[id]/[[view=situationView]]", { id, view }) + search;
}

export function investigationHref(situationId: string, investigationId: string, search = "") {
	return situationHref(situationId, "investigations", investigationSearch(search, investigationId));
}
