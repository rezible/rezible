import type { InvestigationAttributes, SituationAttributes } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";

import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";
import RiCloseCircleLine from "remixicon-svelte/icons/close-circle-line";
import RiDraftLine from "remixicon-svelte/icons/draft-line";
import RiErrorWarningLine from "remixicon-svelte/icons/error-warning-line";
import RiFlagLine from "remixicon-svelte/icons/flag-line";
import RiHistoryLine from "remixicon-svelte/icons/history-line";
import RiEyeLine from "remixicon-svelte/icons/eye-line";
import RiHourglassLine from "remixicon-svelte/icons/hourglass-line";
import RiLoader4Line from "remixicon-svelte/icons/loader-4-line";
import RiPauseCircleLine from "remixicon-svelte/icons/pause-circle-line";
import RiQuestionLine from "remixicon-svelte/icons/question-line";
import RiStopCircleLine from "remixicon-svelte/icons/stop-circle-line";
import RiTimeLine from "remixicon-svelte/icons/time-line";
import RiVolumeMuteLine from "remixicon-svelte/icons/volume-mute-line";

const LONG_RUNNING_AFTER_MS = 24 * 60 * 60 * 1000;

export function closeReasonLabel(reason: SituationAttributes["closeReason"]) {
	switch (reason) {
		case "stabilized":
			return "Stabilized";
		case "expired":
			return "Expired while watching";
		case "merged":
			return "Merged";
		case "dismissed":
			return "Dismissed";
		default:
			return undefined;
	}
}

export function muteReasonLabel(reason: SituationAttributes["muteReason"]) {
	switch (reason) {
		case "not_noteworthy":
			return "Not noteworthy";
		case "expected":
			return "Expected";
		default:
			return undefined;
	}
}

/** Closed, then muted, then raised, otherwise watching. Investigation existence is not a status. */
export function situationStatus(attributes: SituationAttributes): StatusPresentation {
	if (attributes.closedAt) {
		return {
			label: "Closed",
			tone: "neutral",
			icon: RiCheckboxCircleLine,
			description: closeReasonLabel(attributes.closeReason),
		};
	}

	if (attributes.mutedAt) {
		return {
			label: "Muted",
			tone: "neutral",
			icon: RiVolumeMuteLine,
			description: muteReasonLabel(attributes.muteReason),
		};
	}

	return situationStageStatus(attributes.stage);
}

/** Status from the stage alone, for references such as links that carry no close or mute details. */
export function situationStageStatus(stage: SituationAttributes["stage"]): StatusPresentation {
	switch (stage) {
		case "closed":
			return { label: "Closed", tone: "neutral", icon: RiCheckboxCircleLine };
		case "raised":
			return { label: "Raised", tone: "warning", icon: RiFlagLine };
		default:
			return { label: "Watching", tone: "neutral", icon: RiEyeLine };
	}
}

/** Marks an open, unmuted situation raised more than a day ago. */
export function longRunningStatus(
	attributes: SituationAttributes,
	now: number
): StatusPresentation | undefined {
	if (attributes.closedAt || attributes.mutedAt || !attributes.raisedAt) {
		return undefined;
	}

	const raisedAt = Date.parse(attributes.raisedAt);
	if (Number.isNaN(raisedAt) || now - raisedAt <= LONG_RUNNING_AFTER_MS) {
		return undefined;
	}

	return {
		label: "Long-running",
		tone: "warning",
		icon: RiHistoryLine,
		description: "Close it if the noteworthy activity is over",
	};
}

/**
 * The hold deadline while it still protects an open situation. The backend keeps expired deadlines, so a
 * deadline at or before `now` is no hold.
 */
export function activeHoldUntil(attributes: SituationAttributes, now: number): string | undefined {
	if (attributes.closedAt || !attributes.holdUntil) {
		return undefined;
	}
	if (Date.parse(attributes.holdUntil) <= now) {
		return undefined;
	}
	return attributes.holdUntil;
}

export type SituationStateTime = { label: string; at: string; reason?: string };

/** The one time that describes the situation's current state, with the reason for a mute or close. */
export function situationStateTime(attributes: SituationAttributes): SituationStateTime {
	if (attributes.closedAt) {
		return { label: "Closed", at: attributes.closedAt, reason: closeReasonLabel(attributes.closeReason) };
	}
	if (attributes.mutedAt) {
		return { label: "Muted", at: attributes.mutedAt, reason: muteReasonLabel(attributes.muteReason) };
	}
	if (attributes.raisedAt) {
		return { label: "Raised", at: attributes.raisedAt };
	}
	return { label: "Watching since", at: attributes.createdAt };
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

export type InvestigationUpdateNotice = {
	status: StatusPresentation;
	message: string;
	/** Shown after the message, as a time. */
	at?: string;
	/** Pending evidence changes, shown under a failed or stopped headline. */
	pendingMessage?: string;
	offerUpdate: boolean;
};

function evidenceChanges(count: number) {
	if (count === 1) {
		return "1 evidence change";
	}
	return `${count} evidence changes`;
}

function pendingEvidenceMessage(attributes: InvestigationAttributes) {
	const pending = attributes.pendingEvidenceRevisions;
	if (pending <= 0) {
		return undefined;
	}
	if (attributes.automaticUpdatesPaused) {
		return `Automatic updates paused. ${evidenceChanges(pending)} not yet reflected.`;
	}
	return `${evidenceChanges(pending)} not yet reflected.`;
}

/**
 * Whether the investigation's account reflects the current evidence: updating, then a failed or stopped last
 * update, then paused or pending evidence changes, otherwise up to date. Undefined before the first turn.
 * Whenever no turn is queued or running, pending evidence changes are shown and can be updated now.
 */
export function investigationUpdateNotice(
	attributes?: InvestigationAttributes
): InvestigationUpdateNotice | undefined {
	if (!attributes) {
		return undefined;
	}

	const run = investigationRunStatus(attributes);
	if (attributes.activeTurn) {
		return {
			status: run,
			message: "Updating. Findings and the report change as they are published.",
			offerUpdate: false,
		};
	}

	const pendingMessage = pendingEvidenceMessage(attributes);
	const offerUpdate = pendingMessage !== undefined;
	const latestTurn = attributes.latestTurn;
	if (latestTurn?.status === "failed") {
		if (latestTurn.finishedAt) {
			return {
				status: run,
				message: "The last update failed at",
				at: latestTurn.finishedAt,
				pendingMessage,
				offerUpdate,
			};
		}
		return { status: run, message: "The last update failed.", pendingMessage, offerUpdate };
	}
	if (latestTurn?.status === "aborted") {
		return { status: run, message: "The last update was stopped.", pendingMessage, offerUpdate };
	}

	if (pendingMessage) {
		if (attributes.automaticUpdatesPaused) {
			return {
				status: { label: "Paused", tone: "warning", icon: RiPauseCircleLine },
				message: pendingMessage,
				offerUpdate,
			};
		}
		return {
			status: { label: "Pending", tone: "neutral", icon: RiHourglassLine },
			message: pendingMessage,
			offerUpdate,
		};
	}

	if (!latestTurn) {
		return undefined;
	}
	return {
		status: run,
		message: "Up to date with evidence as of",
		at: attributes.evidenceCurrentAsOf,
		offerUpdate: false,
	};
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
