import type { ParamMatcher } from "@sveltejs/kit";

export type SystemViewParam = undefined | "catalogue" | "saved-views";

const views = new Set<string | undefined>([undefined, "catalogue", "saved-views"]);

export const isSystemViewParam = (param: string | undefined): param is SystemViewParam =>
	views.has(param);

export const match = ((param?: string): param is SystemViewParam =>
	isSystemViewParam(param)) satisfies ParamMatcher;
