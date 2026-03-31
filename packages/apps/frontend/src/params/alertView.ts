import type { ParamMatcher } from "@sveltejs/kit";

export type AlertViewParam = undefined | "events" | "incidents" | "playbooks";
const params = new Set([undefined, "events", "incidents", "playbooks"]);
export const match = ((param?: string): param is AlertViewParam => params.has(param)) satisfies ParamMatcher;
