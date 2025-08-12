import type { ParamMatcher } from "@sveltejs/kit";

export type ComponentViewRouteParam = undefined | "incidents";
const params = new Set([undefined, "incidents"]);
export const match = ((param?: string): param is ComponentViewRouteParam => params.has(param)) satisfies ParamMatcher;
