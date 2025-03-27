import type { ParamMatcher } from "@sveltejs/kit";

export type OncallRosterViewRouteParam = undefined | "schedule" | "members" | "backlog" | "analysis" | "resources";
const params = new Set([undefined, "schedule", "members", "backlog", "analysis", "resources"]);

export const convertOncallRosterViewParam = (param?: string): OncallRosterViewRouteParam => {
	if (!params.has(param)) return undefined;
	return param as OncallRosterViewRouteParam;
};

export const match = ((param?: string): param is OncallRosterViewRouteParam => {
	return params.has(param);
}) satisfies ParamMatcher;
