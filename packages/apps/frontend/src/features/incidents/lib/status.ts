import type {
	ExpandableIncidentSeverityAttributes,
	IncidentAttributes,
	RetrospectiveAttributes,
} from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";

import RiAlarmWarningLine from "remixicon-svelte/icons/alarm-warning-line";
import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";
import RiDraftLine from "remixicon-svelte/icons/draft-line";
import RiEditLine from "remixicon-svelte/icons/edit-line";
import RiErrorWarningLine from "remixicon-svelte/icons/error-warning-line";
import RiEyeLine from "remixicon-svelte/icons/eye-line";
import RiFireLine from "remixicon-svelte/icons/fire-line";
import RiInformationLine from "remixicon-svelte/icons/information-line";
import RiQuestionLine from "remixicon-svelte/icons/question-line";
import RiShieldCheckLine from "remixicon-svelte/icons/shield-check-line";
import RiTeamLine from "remixicon-svelte/icons/team-line";

const responseDescription = "Incident response state from the incident tool";
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
		default:
			return {
				label: "Unknown",
				tone: "neutral",
				icon: RiQuestionLine,
				description: responseDescription,
			};
	}
}

export function retrospectiveStatus(state?: RetrospectiveAttributes["state"]): StatusPresentation {
	switch (state) {
		case undefined:
			return {
				label: "Not started",
				tone: "neutral",
				icon: RiDraftLine,
				description: reviewDescription,
			};
		case "draft":
			return { label: "Retrospective", tone: "info", icon: RiEditLine, description: reviewDescription };
		case "in_review":
			return { label: "In review", tone: "warning", icon: RiEyeLine, description: reviewDescription };
		case "meeting":
			return {
				label: "Review meeting",
				tone: "warning",
				icon: RiTeamLine,
				description: reviewDescription,
			};
		case "closed":
			return {
				label: "Closed",
				tone: "success",
				icon: RiCheckboxCircleLine,
				description: reviewDescription,
			};
		default:
			return {
				label: "Unknown",
				tone: "neutral",
				icon: RiQuestionLine,
				description: reviewDescription,
			};
	}
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
