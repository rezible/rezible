import { resolve } from "$app/paths";

/** The catalogue showing one proposal at the top, in any status. */
export function catalogueProposalPath(proposalId: string): string {
	const cataloguePath = resolve("/system/[[view=systemView]]", { view: "catalogue" });
	const query = new URLSearchParams({ proposal: proposalId });
	return `${cataloguePath}?${query.toString()}`;
}
