import type { ParamMatcher } from "@sveltejs/kit";

export type SettingsViewParam = undefined | "integrations" | "incidents";
const params = new Set([undefined, "integrations", "incidents"]);

export const convertSettingsViewParam = (p?: string): SettingsViewParam => {
	return (params.has(p)) ? p as SettingsViewParam : undefined;
};

export const match = ((p?: string): 
	p is SettingsViewParam => (params.has(p))) satisfies ParamMatcher;
