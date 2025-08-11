import type { ParamMatcher } from "@sveltejs/kit";

export type AlertViewRouteParam = undefined | "events" | "playbooks";
const params = new Set([undefined, "events", "playbooks"]);
export const match = ((param?: string): param is AlertViewRouteParam => {
	return params.has(param);
}) satisfies ParamMatcher;
