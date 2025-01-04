import type { ParamMatcher } from '@sveltejs/kit';

const params = new Set([undefined, "timeline", "analysis", "findings"]);
export type IncidentViewRouteParam = undefined | "timeline" | "analysis" | "findings";
export const convertIncidentViewParam = (param?: string): IncidentViewRouteParam => {
	if (!params.has(param)) return undefined;
	return param as IncidentViewRouteParam;
}
export const match = ((param?: string): param is IncidentViewRouteParam => {
	return params.has(param);
}) satisfies ParamMatcher;