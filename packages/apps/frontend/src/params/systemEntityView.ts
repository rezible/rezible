import type { ParamMatcher } from "@sveltejs/kit";

export type SystemEntityViewParam = undefined | "incidents";
const params = new Set([undefined, "incidents"]);
export const match = ((param?: string): param is SystemEntityViewParam => params.has(param)) satisfies ParamMatcher;
