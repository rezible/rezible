import type { ParamMatcher } from "@sveltejs/kit";

export type AlertViewParam = undefined | "events" | "playbooks";
const params = new Set([undefined, "events", "playbooks"]);
export const match = ((param?: string): param is AlertViewParam => params.has(param)) satisfies ParamMatcher;
