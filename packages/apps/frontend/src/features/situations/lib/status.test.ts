import { expect, test } from "bun:test";
import type { InvestigationAttributes, SituationAttributes } from "$lib/api";
import {
	activeHoldUntil,
	hypothesisStatus,
	investigationRunStatus,
	longRunningStatus,
	situationStatus,
} from "./status";

const at = "2026-05-14T04:45:00Z";
const DAY_MS = 24 * 60 * 60 * 1000;

const situation = (attributes: Partial<SituationAttributes>) => attributes as SituationAttributes;
const investigation = (attributes: Partial<InvestigationAttributes>) => attributes as InvestigationAttributes;
const turn = (status: "queued" | "running" | "completed" | "failed" | "aborted") => ({ id: "turn", status });

test("situationStatus precedence: closed, muted, raised, watching", () => {
	const cases: [Partial<SituationAttributes>, string, string, string | undefined][] = [
		[
			{ stage: "closed", closedAt: at, closeReason: "stabilized", mutedAt: at },
			"Closed",
			"neutral",
			"Stabilized",
		],
		[
			{ stage: "closed", closedAt: at, closeReason: "expired" },
			"Closed",
			"neutral",
			"Expired while watching",
		],
		[{ stage: "closed", closedAt: at, closeReason: "merged" }, "Closed", "neutral", "Merged"],
		[{ stage: "closed", closedAt: at, closeReason: "dismissed" }, "Closed", "neutral", "Dismissed"],
		[
			{ stage: "raised", mutedAt: at, muteReason: "not_noteworthy" },
			"Muted",
			"neutral",
			"Not noteworthy",
		],
		[{ stage: "candidate", mutedAt: at, muteReason: "expected" }, "Muted", "neutral", "Expected"],
		[{ stage: "raised", raisedAt: at }, "Raised", "warning", undefined],
		[{ stage: "candidate" }, "Watching", "neutral", undefined],
		[
			{ stage: "candidate", investigation: { investigation: { id: "inv" } } },
			"Watching",
			"neutral",
			undefined,
		],
	];
	for (const [attributes, label, tone, description] of cases) {
		const status = situationStatus(situation(attributes));
		expect([status.label, status.tone, status.description]).toEqual([label, tone, description]);
	}
});

test("longRunningStatus marks only open, unmuted situations raised over a day ago", () => {
	const now = Date.parse(at);
	const dayAgo = new Date(now - DAY_MS).toISOString();
	const overDayAgo = new Date(now - DAY_MS - 1).toISOString();

	expect(longRunningStatus(situation({ stage: "raised", raisedAt: overDayAgo }), now)?.label).toBe(
		"Long-running"
	);
	expect(longRunningStatus(situation({ stage: "raised", raisedAt: dayAgo }), now)).toBeUndefined();
	expect(longRunningStatus(situation({ stage: "candidate" }), now)).toBeUndefined();
	expect(
		longRunningStatus(situation({ stage: "raised", raisedAt: overDayAgo, mutedAt: at }), now)
	).toBeUndefined();
	expect(
		longRunningStatus(situation({ stage: "closed", raisedAt: overDayAgo, closedAt: at }), now)
	).toBeUndefined();
});

test("activeHoldUntil protects only before the deadline", () => {
	const now = Date.parse(at);
	const later = new Date(now + 1).toISOString();

	expect(activeHoldUntil(situation({ holdUntil: later }), now)).toBe(later);
	expect(activeHoldUntil(situation({ holdUntil: at }), now)).toBeUndefined();
	expect(activeHoldUntil(situation({ holdUntil: new Date(now - 1).toISOString() }), now)).toBeUndefined();
	expect(activeHoldUntil(situation({}), now)).toBeUndefined();
	expect(activeHoldUntil(situation({ holdUntil: later, closedAt: at }), now)).toBeUndefined();
});

test("investigationRunStatus prefers the active turn", () => {
	const cases: [Partial<InvestigationAttributes> | undefined, string, string][] = [
		[{ activeTurn: turn("queued"), latestTurn: turn("failed") }, "Queued", "neutral"],
		[{ activeTurn: turn("running") }, "Running", "info"],
		[{ activeTurn: null, latestTurn: turn("completed") }, "Completed", "success"],
		[{ activeTurn: null, latestTurn: turn("failed") }, "Failed", "danger"],
		[{ activeTurn: null, latestTurn: turn("aborted") }, "Stopped", "neutral"],
		[{ activeTurn: null, latestTurn: null }, "Not started", "neutral"],
		[undefined, "Not started", "neutral"],
	];
	for (const [attributes, label, tone] of cases) {
		const status = investigationRunStatus(attributes ? investigation(attributes) : undefined);
		expect([status.label, status.tone]).toEqual([label, tone]);
	}
	expect(investigationRunStatus(investigation({ latestTurn: turn("failed") })).description).toBe(
		"The investigation turn could not finish."
	);
});

test("hypothesisStatus maps every status and falls back to Unknown", () => {
	const cases: [string, string, string][] = [
		["open", "Open", "neutral"],
		["supported", "Supported", "success"],
		["disproven", "Disproven", "neutral"],
		["inconclusive", "Inconclusive", "warning"],
		["something-else", "Unknown", "neutral"],
	];
	for (const [value, label, tone] of cases) {
		const status = hypothesisStatus(value);
		expect([status.label, status.tone]).toEqual([label, tone]);
	}
});
