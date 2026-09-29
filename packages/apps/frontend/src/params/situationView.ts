import type { ParamMatcher } from "@sveltejs/kit";

export type SituationViewParam = undefined | "investigation";
const values = new Set<string | undefined>([undefined, "investigation"]);
export const match = ((param?: string): param is SituationViewParam =>
	values.has(param)) satisfies ParamMatcher;
