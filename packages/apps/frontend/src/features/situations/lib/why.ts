import type { SituationAttributes, SituationJudgment, SituationReason } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { closeReasonLabel, muteReasonLabel } from "./status";
import RiErrorWarningLine from "remixicon-svelte/icons/error-warning-line";
import RiFlagLine from "remixicon-svelte/icons/flag-line";
import RiForbidLine from "remixicon-svelte/icons/forbid-line";
import RiPauseCircleLine from "remixicon-svelte/icons/pause-circle-line";

type ReasonKey = SituationReason["reason"];

const reasonOrder: { key: ReasonKey; label: string }[] = [
	{ key: "linked_incident", label: "Linked incident" },
	{ key: "breadth", label: "Breadth" },
	{ key: "novelty", label: "Novelty" },
	{ key: "persistence", label: "Persistence" },
	{ key: "past_incident", label: "Past incident" },
];

const unavailableJudge = "unavailable";
const rulesJudge = "rules";
const rejectedJudgeSuffix = ":rejected";
const modelJudgePrefix = "llm:";

/** The seed signal's details, from the situation's alert episodes. */
export type WhyOrigin = { title: string; startedAt: string };

/** One reason as recorded in the judgment, not a current checklist. */
export type WhyReason = { key: ReasonKey; label: string; met: boolean; cited: boolean; detail: string };

export type WhyAssessment = {
	/** "Earlier assessment" when something other than this assessment raised the situation. */
	title: string;
	judgedAt: string;
	verdict: StatusPresentation;
	/** Shown in place of a confident decision when the judge could not answer or its answer was rejected. */
	notice?: string;
	reasons: WhyReason[];
	/** The stored explanation, verbatim. */
	explanation: string;
	/** Frames a stored retry promise that no longer applies because the situation is no longer a candidate. */
	explanationContext?: string;
	assessedBy: string;
	/** True only for an unmuted, unraised, unclosed candidate, which Rezible keeps assessing. */
	checksAgain: boolean;
};

export type WhyRaise = {
	by: "person" | "rezible" | "system";
	at: string;
	/** The recorded reason of the first raised action, when it has one. */
	reason?: string;
};

export type WhyClosure = {
	at: string;
	reasonLabel?: string;
	/** The `merged_into` target of a merged source; its signals moved there. */
	mergedInto?: { id: string; title: string };
};

type WhyBase = { heading: string; origin?: WhyOrigin };

export type WhyPanelState =
	| (WhyBase & { kind: "pending" })
	| (WhyBase & { kind: "assessed"; assessment: WhyAssessment; raise?: WhyRaise; closure?: WhyClosure })
	| (WhyBase & { kind: "raised"; raise: WhyRaise; earlierAssessment?: WhyAssessment })
	| (WhyBase & { kind: "muted"; mutedAt: string; reasonLabel?: string; lastAssessedAt?: string })
	| (WhyBase & { kind: "closed-unassessed"; closure: WhyClosure; raise?: WhyRaise });

const raiseVerdict: StatusPresentation = { label: "Raise", tone: "warning", icon: RiFlagLine };
const holdVerdict: StatusPresentation = { label: "Hold", tone: "neutral", icon: RiPauseCircleLine };
const unavailableVerdict: StatusPresentation = {
	label: "Not assessed",
	tone: "warning",
	icon: RiErrorWarningLine,
};
const rejectedVerdict: StatusPresentation = { label: "Answer rejected", tone: "warning", icon: RiForbidLine };

/**
 * The "why" panel's state, from the stored latest judgment and the first raised action.
 * Precedence: closed, then muted, then raised, otherwise watching. Performs no reads.
 */
export function whyPanelState(attributes: SituationAttributes, origin?: WhyOrigin): WhyPanelState {
	const judgment = attributes.latestJudgment;
	const raise = whyRaise(attributes);

	if (attributes.closedAt) {
		const closure = whyClosure(attributes, attributes.closedAt);
		const heading = "Why Rezible watched this";
		if (!judgment) {
			return { kind: "closed-unassessed", heading, origin, closure, raise };
		}
		const assessment = whyAssessment(judgment, assessmentTitle(raise), "closed");
		return { kind: "assessed", heading, origin, assessment, raise, closure };
	}

	if (attributes.mutedAt) {
		return {
			kind: "muted",
			heading: "Muted",
			origin,
			mutedAt: attributes.mutedAt,
			reasonLabel: muteReasonLabel(attributes.muteReason),
			lastAssessedAt: judgment?.judgedAt,
		};
	}

	if (raise) {
		if (raise.by === "rezible" && judgment) {
			const assessment = whyAssessment(judgment, "Assessment", "raised");
			return { kind: "assessed", heading: "Why this was raised", origin, assessment, raise };
		}

		let earlierAssessment: WhyAssessment | undefined;
		if (judgment) {
			earlierAssessment = whyAssessment(judgment, "Earlier assessment", "raised");
		}
		let heading = "Raised";
		if (raise.by === "person") {
			heading = "Raised by a person";
		}
		return { kind: "raised", heading, origin, raise, earlierAssessment };
	}

	const heading = "Why Rezible is watching this";
	if (!judgment) {
		return { kind: "pending", heading, origin };
	}
	const assessment = whyAssessment(judgment, "Assessment", "candidate");
	return { kind: "assessed", heading, origin, assessment };
}

/**
 * Who raised the situation, from its first raised action. Rezible raised it only when no person acted and
 * the latest judgment decided to raise; any other raise without a person is another system path.
 */
function whyRaise(attributes: SituationAttributes): WhyRaise | undefined {
	if (!attributes.raisedAt) {
		return undefined;
	}

	const raisedBy = attributes.raisedBy;
	const reason = raisedBy?.reason.trim() || undefined;
	if (raisedBy?.userId) {
		return { by: "person", at: attributes.raisedAt, reason };
	}
	if (attributes.latestJudgment?.decision === "raise") {
		// The action's reason is the cited reason keys; the assessment presents them.
		return { by: "rezible", at: attributes.raisedAt };
	}
	return { by: "system", at: attributes.raisedAt, reason };
}

function whyClosure(attributes: SituationAttributes, closedAt: string): WhyClosure {
	const closure: WhyClosure = { at: closedAt, reasonLabel: closeReasonLabel(attributes.closeReason) };
	if (attributes.closeReason === "merged") {
		const target = attributes.links.find((link) => link.kind === "merged_into");
		if (target) {
			closure.mergedInto = { id: target.situationId, title: target.title };
		}
	}
	return closure;
}

/** A judgment that did not raise the situation is never presented as the cause of a raise. */
function assessmentTitle(raise: WhyRaise | undefined) {
	if (raise && raise.by !== "rezible") {
		return "Earlier assessment";
	}
	return "Assessment";
}

type Standing = "candidate" | "raised" | "closed";

function whyAssessment(judgment: SituationJudgment, title: string, standing: Standing): WhyAssessment {
	const assessment: WhyAssessment = {
		title,
		judgedAt: judgment.judgedAt,
		verdict: decisionVerdict(judgment.decision),
		reasons: whyReasons(judgment),
		explanation: judgment.explanation.trim(),
		assessedBy: assessedBy(judgment.judge),
		checksAgain: standing === "candidate",
	};

	if (judgment.judge === unavailableJudge) {
		assessment.verdict = unavailableVerdict;
		if (standing === "candidate") {
			assessment.notice = "Rezible could not assess this; it will retry.";
		} else {
			assessment.notice = "Rezible could not assess this at the time.";
			assessment.explanationContext = staleRetryContext(standing);
		}
	} else if (judgment.judge.endsWith(rejectedJudgeSuffix)) {
		assessment.verdict = rejectedVerdict;
		assessment.notice = "Rezible's answer was rejected, so no raise was made.";
	}

	return assessment;
}

/** An unavailable judgment promises a retry; that promise ended when the situation stopped being a candidate. */
function staleRetryContext(standing: "raised" | "closed") {
	if (standing === "closed") {
		return "Recorded while the judge was unavailable. The retry it mentions did not apply once the situation closed.";
	}
	return "Recorded while the judge was unavailable. The retry it mentions did not apply once the situation was raised.";
}

function decisionVerdict(decision: SituationJudgment["decision"]) {
	if (decision === "raise") {
		return raiseVerdict;
	}
	return holdVerdict;
}

function whyReasons(judgment: SituationJudgment): WhyReason[] {
	const cited = new Set(judgment.citedReasons);
	return reasonOrder.map(({ key, label }) => {
		const stored = judgment.reasons.find((reason) => reason.reason === key);
		if (!stored) {
			return { key, label, met: false, cited: false, detail: "Not recorded." };
		}
		return { key, label, met: stored.met, cited: cited.has(key), detail: stored.detail };
	});
}

function assessedBy(judge: string) {
	if (judge === rulesJudge) {
		return "Assessed by the rules";
	}
	if (judge === unavailableJudge || judge.endsWith(rejectedJudgeSuffix)) {
		return "Asked Rezible's judgment";
	}
	if (judge.startsWith(modelJudgePrefix)) {
		return "Assessed by Rezible's judgment";
	}
	return "Assessed by Rezible";
}
