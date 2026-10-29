import type { Incident } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { incidentServiceImpacts } from "./impacts";
import { incidentResponseStatus, incidentSeverityStatus } from "./status";

export type IncidentListRow = {
	title: string;
	reference: string;
	summary: string;
	response: StatusPresentation;
	/** Undefined when the incident has no severity. */
	severity: StatusPresentation | undefined;
	openedAt: string;
	updatedAt: string;
	services: string[];
	moreServices: number;
};

const MAX_SERVICES = 3;

export function incidentListRow(incident: Incident): IncidentListRow {
	const attributes = incident.attributes;
	const names = [...new Set(incidentServiceImpacts(attributes).map((service) => service.name))];

	let severity: StatusPresentation | undefined;
	if (attributes.severity?.attributes?.name) {
		severity = incidentSeverityStatus(attributes.severity);
	}

	return {
		title: attributes.title,
		reference: attributes.reference,
		summary: attributes.summary.trim(),
		response: incidentResponseStatus(attributes.responseState),
		severity,
		openedAt: attributes.openedAt,
		updatedAt: attributes.updatedAt,
		services: names.slice(0, MAX_SERVICES),
		moreServices: Math.max(0, names.length - MAX_SERVICES),
	};
}
