import type { ParamMatcher } from "@sveltejs/kit";

export type SystemTopologyEntityViewParam = undefined | "incidents";
const params = new Set([undefined, "incidents"]);
export const match = ((param?: string): param is SystemTopologyEntityViewParam => params.has(param)) satisfies ParamMatcher;
