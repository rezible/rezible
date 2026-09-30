import type { InvestigationAttributes, SituationAttributes } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";

import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";
import RiCloseCircleLine from "remixicon-svelte/icons/close-circle-line";
import RiDraftLine from "remixicon-svelte/icons/draft-line";
import RiErrorWarningLine from "remixicon-svelte/icons/error-warning-line";
import RiEyeLine from "remixicon-svelte/icons/eye-line";
import RiHourglassLine from "remixicon-svelte/icons/hourglass-line";
import RiLoader4Line from "remixicon-svelte/icons/loader-4-line";
import RiQuestionLine from "remixicon-svelte/icons/question-line";
import RiSearchLine from "remixicon-svelte/icons/search-line";
import RiStopCircleLine from "remixicon-svelte/icons/stop-circle-line";
import RiTimeLine from "remixicon-svelte/icons/time-line";

function closeReasonLabel(reason: SituationAttributes["closeReason"]) {
	switch (reason) {
		case "stabilized":
			return "Stabilized";
		case "dismissed":
			return "Dismissed";
		default:
			return undefined;
	}
}

export function situationStatus(attributes: SituationAttributes): StatusPresentation {
	if (attributes.closedAt) {
		return {
			label: "Closed",
			tone: "neutral",
			icon: RiCheckboxCircleLine,
			description: closeReasonLabel(attributes.closeReason),
		};
	}

	if (attributes.investigation) {
		return { label: "Investigating", tone: "warning", icon: RiSearchLine };
	}

	return { label: "Observed", tone: "info", icon: RiEyeLine };
}

export function investigationRunStatus(attributes?: InvestigationAttributes): StatusPresentation {
	switch (attributes?.activeTurn?.status) {
		case "queued":
			return { label: "Queued", tone: "neutral", icon: RiTimeLine };
		case "running":
			return { label: "Running", tone: "info", icon: RiLoader4Line };
	}

	switch (attributes?.latestTurn?.status) {
		case "completed":
			return { label: "Completed", tone: "success", icon: RiCheckboxCircleLine };
		case "failed":
			return {
				label: "Failed",
				tone: "danger",
				icon: RiErrorWarningLine,
				description: "The investigation turn could not finish.",
			};
		case "aborted":
			return {
				label: "Stopped",
				tone: "neutral",
				icon: RiStopCircleLine,
				description: "The investigation turn was stopped.",
			};
		default:
			return { label: "Not started", tone: "neutral", icon: RiTimeLine };
	}
}

export function hypothesisStatus(status: string): StatusPresentation {
	switch (status) {
		case "open":
			return { label: "Open", tone: "neutral", icon: RiQuestionLine };
		case "supported":
			return { label: "Supported", tone: "success", icon: RiCheckboxCircleLine };
		case "disproven":
			return { label: "Disproven", tone: "neutral", icon: RiCloseCircleLine };
		case "inconclusive":
			return { label: "Inconclusive", tone: "warning", icon: RiQuestionLine };
		default:
			return { label: "Unknown", tone: "neutral", icon: RiQuestionLine };
	}
}

export const provisionalStatus: StatusPresentation = {
	label: "Provisional",
	tone: "info",
	icon: RiDraftLine,
	description: "Published by an investigation turn that has not finished",
};

export const followUpPendingStatus: StatusPresentation = {
	label: "Follow-up pending",
	tone: "warning",
	icon: RiHourglassLine,
};

export const invalidatedStatus: StatusPresentation = {
	label: "Invalidated",
	tone: "neutral",
	icon: RiCloseCircleLine,
	description: "A later finding invalidates this one.",
};
