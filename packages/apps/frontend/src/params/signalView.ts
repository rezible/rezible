import type { ParamMatcher } from "@sveltejs/kit";

export type SignalViewParam = undefined | "events" | "incidents";
const params = new Set([undefined, "events", "incidents"]);
export const match = ((param?: string): param is SignalViewParam => params.has(param)) satisfies ParamMatcher;
