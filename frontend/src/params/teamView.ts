import type { ParamMatcher } from "@sveltejs/kit";

export type TeamViewRouteParam = undefined | "backlog";
const params = new Set([undefined, "backlog"]);
export const convertTeamViewParam = (param?: string): TeamViewRouteParam => {
	if (!params.has(param)) return undefined;
	return param as TeamViewRouteParam;
};
export const match = ((param?: string): param is TeamViewRouteParam => params.has(param)) satisfies ParamMatcher;
