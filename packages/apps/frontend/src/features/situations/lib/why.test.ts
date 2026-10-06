import { expect, test } from "bun:test";
import type { SituationAttributes, SituationJudgment } from "$lib/api";
import { whyPanelState } from "./why";

const openedAt = "2026-05-14T04:00:00Z";
const judgedAt = "2026-05-14T04:10:00Z";
const raisedAt = "2026-05-14T04:20:00Z";
const closedAt = "2026-05-14T05:00:00Z";
const origin = { title: "High error rate", startedAt: openedAt };

const situation = (attributes: Partial<SituationAttributes>) =>
	({ stage: "candidate", links: [], ...attributes }) as SituationAttributes;

const judgment = (attributes: Partial<SituationJudgment>): SituationJudgment => ({
	judgedAt,
	outcome: "needs_decision",
	decision: "hold",
	reasons: [
		{ reason: "breadth", met: true, hard: false, detail: "3 services" },
		{ reason: "novelty", met: false, hard: false, detail: "Seen weekly" },
	],
	citedReasons: [],
	explanation: "Held: only one source.",
	judge: "llm:judge_situation_candidate",
	...attributes,
});

test("a candidate without a judgment is pending", () => {
	const why = whyPanelState(situation({}), origin);
	expect(why.kind).toBe("pending");
	expect(why.heading).toBe("Why Rezible is watching this");
	expect(why.origin).toEqual(origin);
});

test("a candidate with a judgment shows the last assessment and keeps checking", () => {
	const why = whyPanelState(situation({ latestJudgment: judgment({}) }));
	if (why.kind !== "assessed") {
		throw new Error(why.kind);
	}
	expect(why.heading).toBe("Why Rezible is watching this");
	expect(why.assessment.checksAgain).toBe(true);
	expect(why.assessment.reasons.map((reason) => [reason.key, reason.met])).toEqual([
		["linked_incident", false],
		["breadth", true],
		["novelty", false],
		["persistence", false],
		["past_incident", false],
	]);
});

test("an automatic raise shows the recorded decision", () => {
	const why = whyPanelState(
		situation({
			stage: "raised",
			raisedAt,
			raisedBy: { reason: "breadth" },
			latestJudgment: judgment({ decision: "raise", citedReasons: ["breadth"] }),
		}),
		origin
	);
	if (why.kind !== "assessed") {
		throw new Error(why.kind);
	}
	expect(why.heading).toBe("Why this was raised");
	expect(why.raise?.by).toBe("rezible");
	expect(why.assessment.title).toBe("Assessment");
	expect(why.assessment.checksAgain).toBe(false);
	expect(why.assessment.reasons.find((reason) => reason.key === "breadth")?.cited).toBe(true);
});

test("a person's raise after a hold is not attributed to the hold", () => {
	const why = whyPanelState(
		situation({
			stage: "raised",
			raisedAt,
			raisedBy: { userId: "user", reason: "Customer reports" },
			latestJudgment: judgment({}),
		})
	);
	if (why.kind !== "raised") {
		throw new Error(why.kind);
	}
	expect(why.heading).toBe("Raised by a person");
	expect(why.raise).toEqual({ by: "person", at: raisedAt, reason: "Customer reports" });
	expect(why.earlierAssessment?.title).toBe("Earlier assessment");
});

test("a raise through another system path never names a person", () => {
	const why = whyPanelState(
		situation({ stage: "raised", raisedAt, raisedBy: { reason: "linked incident" } })
	);
	if (why.kind !== "raised") {
		throw new Error(why.kind);
	}
	expect(why.heading).toBe("Raised");
	expect(why.raise.by).toBe("system");
	expect(why.raise.reason).toBe("linked incident");
});

test("a muted situation is not being assessed", () => {
	const why = whyPanelState(
		situation({ mutedAt: closedAt, muteReason: "expected", latestJudgment: judgment({}) })
	);
	expect(why).toMatchObject({ kind: "muted", reasonLabel: "Expected", lastAssessedAt: judgedAt });
});

test("a closed situation with a judgment shows the stored assessment with the closure", () => {
	const why = whyPanelState(
		situation({ stage: "closed", closedAt, closeReason: "expired", latestJudgment: judgment({}) }),
		origin
	);
	if (why.kind !== "assessed") {
		throw new Error(why.kind);
	}
	expect(why.closure).toEqual({ at: closedAt, reasonLabel: "Expired while watching" });
	expect(why.assessment.judgedAt).toBe(judgedAt);
	expect(why.assessment.checksAgain).toBe(false);
});

test("a closed situation without a judgment shows origin and closure", () => {
	const why = whyPanelState(situation({ stage: "closed", closedAt, closeReason: "dismissed" }), origin);
	expect(why).toMatchObject({ kind: "closed-unassessed", origin, closure: { reasonLabel: "Dismissed" } });
});

test("a merged source links to its target with or without a judgment and without origin", () => {
	const merged = {
		stage: "closed" as const,
		closedAt,
		closeReason: "merged" as const,
		links: [
			{
				kind: "merged_into" as const,
				situationId: "target",
				title: "Checkout outage",
				stage: "raised" as const,
			},
		],
	};
	const mergedInto = { id: "target", title: "Checkout outage" };

	const unassessed = whyPanelState(situation(merged));
	expect(unassessed).toMatchObject({ kind: "closed-unassessed", closure: { mergedInto } });
	expect(unassessed.origin).toBeUndefined();

	const assessed = whyPanelState(situation({ ...merged, latestJudgment: judgment({}) }));
	expect(assessed).toMatchObject({ kind: "assessed", closure: { mergedInto } });
	expect(assessed.origin).toBeUndefined();
});

test("an unavailable judge promises a retry only while watching", () => {
	const unavailable = judgment({ judge: "unavailable" });

	const watching = whyPanelState(situation({ latestJudgment: unavailable }));
	if (watching.kind !== "assessed") {
		throw new Error(watching.kind);
	}
	expect(watching.assessment.verdict.label).toBe("Not assessed");
	expect(watching.assessment.explanationContext).toBeUndefined();

	const raised = whyPanelState(
		situation({
			stage: "raised",
			raisedAt,
			raisedBy: { userId: "user", reason: "" },
			latestJudgment: unavailable,
		})
	);
	if (raised.kind !== "raised" || !raised.earlierAssessment) {
		throw new Error(raised.kind);
	}
	expect(raised.earlierAssessment.verdict.label).toBe("Not assessed");
	expect(raised.earlierAssessment.explanationContext).toBeDefined();
});

test("a rejected answer is not a confident assessment", () => {
	const why = whyPanelState(
		situation({ latestJudgment: judgment({ judge: "llm:judge_situation_candidate:rejected" }) })
	);
	if (why.kind !== "assessed") {
		throw new Error(why.kind);
	}
	expect(why.assessment.verdict.label).toBe("Answer rejected");
	expect(why.assessment.notice).toBeDefined();
});
