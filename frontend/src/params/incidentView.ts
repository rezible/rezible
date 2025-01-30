import type { ParamMatcher } from "@sveltejs/kit";

const params = new Set([undefined, "analysis", "report"]);
export type IncidentViewRouteParam = undefined | "analysis" | "report";
export const convertIncidentViewParam = (
  param?: string,
): IncidentViewRouteParam => {
  if (!params.has(param)) return undefined;
  return param as IncidentViewRouteParam;
};
export const match = ((param?: string): param is IncidentViewRouteParam => {
  return params.has(param);
}) satisfies ParamMatcher;
