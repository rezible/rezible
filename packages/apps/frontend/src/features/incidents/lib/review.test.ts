import { expect, test } from "bun:test";
import type { RetrospectiveAttributes } from "$lib/api";
import { closedReviewLine, reviewPanelModel } from "./review";

const retrospective: RetrospectiveAttributes = {
	documentId: "doc",
	systemAnalysisId: "analysis",
	reportSections: [],
	state: "in_review",
	owner: {
		id: "owner",
		attributes: { name: "Owner", email: "owner@example.com", organizationRole: "member" },
	},
	reviews: [
		{
			reviewer: { id: "reviewer" },
			state: "waiting",
			updatedAt: "2026-10-24T00:00:00Z",
			messageFailed: false,
		},
	],
};

test("review actions follow ownership, reviewer membership and pending decisions", () => {
	const owner = reviewPanelModel(retrospective, "owner");
	expect(owner.actions.changeReviewers).toBe(true);
	expect(owner.actions.canClose).toBe(false);
	expect(owner.actions.decide).toBe(false);
	expect(reviewPanelModel(retrospective, "reviewer").actions.decide).toBe(true);
	expect(Object.values(reviewPanelModel(retrospective, "other").actions).some(Boolean)).toBe(false);
	const approved = {
		...retrospective,
		reviews: retrospective.reviews.map((r) => ({ ...r, state: "approved" as const })),
	};
	expect(reviewPanelModel(approved, "owner").actions.canClose).toBe(true);
});

test("closed reviews have no actions and preserve the closure outcome", () => {
	for (const outcome of ["reviewed", "closed_without_review", "skipped"] as const) {
		const model = reviewPanelModel({ ...retrospective, state: "closed", outcome }, "owner");
		expect(Object.values(model.actions).some(Boolean)).toBe(false);
		expect(model.closed).toBe(true);
	}
});

test("the closed banner line names the closer, time and outcome", () => {
	const closed: RetrospectiveAttributes = {
		...retrospective,
		state: "closed",
		closedBy: {
			id: "ava",
			attributes: { name: "Ava Patel", email: "ava@example.com", organizationRole: "member" },
		},
		closedAt: new Date(2026, 9, 6, 14, 32).toISOString(),
		skipReason: "Duplicate of INC-12",
	};
	expect(closedReviewLine({ ...closed, outcome: "reviewed" })).toBe(
		"Closed by Ava Patel after review · Tue 6 Oct, 14:32"
	);
	expect(closedReviewLine({ ...closed, outcome: "closed_without_review" })).toBe(
		"Closed without review by Ava Patel · Tue 6 Oct, 14:32"
	);
	expect(closedReviewLine({ ...closed, outcome: "skipped" })).toBe(
		"Skipped by Ava Patel · Tue 6 Oct, 14:32: Duplicate of INC-12"
	);
});
