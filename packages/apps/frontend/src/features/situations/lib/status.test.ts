import { expect, test } from "bun:test";
import type { InvestigationAttributes, SituationAttributes } from "$lib/api";
import { hypothesisStatus, investigationRunStatus, situationStatus } from "./status";

const situation = (attributes: Partial<SituationAttributes>) => attributes as SituationAttributes;
const investigation = (attributes: Partial<InvestigationAttributes>) => attributes as InvestigationAttributes;
const turn = (status: "queued" | "running" | "completed" | "failed" | "aborted") => ({ id: "turn", status });

test("situationStatus precedence", () => {
	const closed = situationStatus(
		situation({ closedAt: "2026-05-14T04:45:00Z", closeReason: "stabilized" })
	);
	expect([closed.label, closed.tone, closed.description]).toEqual(["Closed", "neutral", "Stabilized"]);

	const dismissed = situationStatus(
		situation({ closedAt: "2026-05-14T04:45:00Z", closeReason: "dismissed" })
	);
	expect(dismissed.description).toBe("Dismissed");

	const investigating = situationStatus(situation({ investigation: { investigation: { id: "inv" } } }));
	expect([investigating.label, investigating.tone]).toEqual(["Investigating", "warning"]);

	const observed = situationStatus(situation({}));
	expect([observed.label, observed.tone]).toEqual(["Observed", "info"]);
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
