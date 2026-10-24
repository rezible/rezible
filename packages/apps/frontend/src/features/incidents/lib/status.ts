import { reviewPanelModel } from "./review";
import type {
	ExpandableIncidentSeverityAttributes,
	IncidentAttributes,
	RetrospectiveAttributes,
} from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";

import RiAlarmWarningLine from "remixicon-svelte/icons/alarm-warning-line";
import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";
import RiErrorWarningLine from "remixicon-svelte/icons/error-warning-line";
import RiFireLine from "remixicon-svelte/icons/fire-line";
import RiInformationLine from "remixicon-svelte/icons/information-line";
import RiShieldCheckLine from "remixicon-svelte/icons/shield-check-line";

const responseDescription = "Incident response state, from its milestones";
const reviewDescription = "Rezible review stage";

export function incidentResponseStatus(state: IncidentAttributes["responseState"]): StatusPresentation {
	switch (state) {
		case "started":
			return { label: "Active", tone: "danger", icon: RiFireLine, description: responseDescription };
		case "mitigated":
			return {
				label: "Mitigated",
				tone: "warning",
				icon: RiShieldCheckLine,
				description: responseDescription,
			};
		case "resolved":
			return {
				label: "Resolved",
				tone: "success",
				icon: RiCheckboxCircleLine,
				description: responseDescription,
			};
	}
}

export function retrospectiveStatus(
	state?: RetrospectiveAttributes["state"],
	outcome?: RetrospectiveAttributes["outcome"]
): StatusPresentation {
	return reviewPanelModel(
		state
			? { state, outcome, reviews: [], documentId: "", systemAnalysisId: "", reportSections: [] }
			: undefined
	).status;
}

export function incidentSeverityStatus(severity?: ExpandableIncidentSeverityAttributes): StatusPresentation {
	const name = severity?.attributes?.name;
	if (!name) {
		return { label: "No severity", tone: "neutral", icon: RiInformationLine };
	}

	const rank = severity?.attributes?.rank ?? Number.POSITIVE_INFINITY;
	if (rank <= 1) {
		return { label: name, tone: "danger", icon: RiAlarmWarningLine };
	}
	if (rank === 2) {
		return { label: name, tone: "warning", icon: RiErrorWarningLine };
	}
	return { label: name, tone: "neutral", icon: RiInformationLine };
}
