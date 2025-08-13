import type { ParamMatcher } from "@sveltejs/kit";

export type ComponentViewParam = undefined | "incidents";
const params = new Set([undefined, "incidents"]);
export const match = ((param?: string): param is ComponentViewParam => params.has(param)) satisfies ParamMatcher;
