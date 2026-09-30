import type { IncidentAttributes } from "$lib/api";

export type IncidentServiceImpact = {
	key: string;
	name: string;
	note: string | undefined;
	source: string | undefined;
};

function trimmedOrUndefined(value: string | undefined) {
	const trimmed = value?.trim();
	if (trimmed) {
		return trimmed;
	}
	return undefined;
}

/** Impacts whose knowledge entity has a display name, in API order. */
export function incidentServiceImpacts(attributes: IncidentAttributes): IncidentServiceImpact[] {
	const services: IncidentServiceImpact[] = [];
	for (const impact of attributes.impacts) {
		const name = impact.knowledgeEntity.attributes?.latestState?.displayName;
		if (!name) {
			continue;
		}
		services.push({
			key: impact.id,
			name,
			note: trimmedOrUndefined(impact.note),
			source: trimmedOrUndefined(impact.source),
		});
	}
	return services;
}
