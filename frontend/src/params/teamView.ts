import type { ParamMatcher } from "@sveltejs/kit";

export type TeamViewParam = undefined | "backlog" | "meetings";
const params = new Set([undefined, "backlog", "meetings"]);
export const match = ((param?: string): param is TeamViewParam => params.has(param)) satisfies ParamMatcher;
