import type { ParamMatcher } from "@sveltejs/kit";

export type SituationViewParam = undefined | "impact" | "investigations";
const values = new Set<string | undefined>([undefined, "impact", "investigations"]);
export const match = ((param?: string): param is SituationViewParam =>
	values.has(param)) satisfies ParamMatcher;
