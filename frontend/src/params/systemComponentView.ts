import type { ParamMatcher } from "@sveltejs/kit";

export type SystemComponentViewParam = undefined | "incidents";
const params = new Set([undefined, "incidents"]);
export const match = ((param?: string): param is SystemComponentViewParam => params.has(param)) satisfies ParamMatcher;
