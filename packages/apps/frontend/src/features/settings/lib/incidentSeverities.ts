import type { IncidentSeverity } from "$lib/api";

// Lower ranks are more severe and read first. Equal ranks read in name order.
export function sortSeveritiesByRank(severities: IncidentSeverity[]) {
	return severities.toSorted((a, b) => {
		const rankOrder = a.attributes.rank - b.attributes.rank;
		if (rankOrder !== 0) {
			return rankOrder;
		}

		return a.attributes.name.localeCompare(b.attributes.name);
	});
}
