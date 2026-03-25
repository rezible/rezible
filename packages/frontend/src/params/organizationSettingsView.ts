import type { ParamMatcher } from "@sveltejs/kit";

export type OrganizationSettingsViewParam = undefined | "users";
const params = new Set([undefined, "users"]);

export const convertOrganizationSettingsViewParam = (param?: string): OrganizationSettingsViewParam => {
	if (!params.has(param)) return undefined;
	return param as OrganizationSettingsViewParam;
};

export const match = ((param?: string): param is OrganizationSettingsViewParam => {
	return params.has(param);
}) satisfies ParamMatcher;
