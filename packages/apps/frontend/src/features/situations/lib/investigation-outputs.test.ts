import { expect, test } from "bun:test";
import type { InvestigationFinding, InvestigationHypothesis, InvestigationReportAttributes } from "$lib/api";
import {
	buildCitationIndex,
	buildFindingViews,
	buildHypothesisViews,
	findingRow,
	hypothesisRow,
} from "./investigation-outputs";

const ref = (id: string) => ({ id, kind: "knowledge_evidence" as const });

const finding = (
	id: string,
	overrides: Partial<InvestigationFinding["attributes"]> = {}
): InvestigationFinding => ({
	id,
	attributes: {
		agentTurnId: "turn",
		body: `Body of ${id}.`,
		createdAt: "2026-05-14T04:45:00Z",
		findingId: `finding-${id}`,
		findingReferences: [],
		invalidatedByVersionIds: [],
		key: id,
		provisional: false,
		references: [],
		title: `Finding ${id}`,
		turnStatus: "completed",
		userInputId: null,
		...overrides,
	},
});

const hypothesis = (
	id: string,
	overrides: Partial<InvestigationHypothesis["attributes"]> = {}
): InvestigationHypothesis => ({
	id,
	attributes: {
		agentTurnId: "turn",
		createdAt: "2026-05-14T04:45:00Z",
		hypothesisId: `hypothesis-${id}`,
		justification: "Justification.",
		key: id,
		provisional: false,
		references: [],
		status: "open",
		title: `Hypothesis ${id}`,
		turnStatus: "completed",
		...overrides,
	},
});

const report = (references: string[]): InvestigationReportAttributes => ({
	agentTurnId: "turn",
	createdAt: "2026-05-14T04:45:00Z",
	provisional: false,
	references: references.map(ref),
	summary: "",
	text: "Report.",
	turnStatus: "completed",
});

test("answers are excluded and findings are numbered contiguously", () => {
	const findings = [finding("a"), finding("answer", { userInputId: "input" }), finding("b")];
	const views = buildFindingViews(findings, new Map());
	expect(views.map((view) => [view.id, view.number])).toEqual([
		["a", 1],
		["b", 2],
	]);
});

test("citation numbers follow report, findings, then hypotheses", () => {
	const findings = [
		finding("a", { references: [ref("e2"), ref("e3")] }),
		finding("answer", { userInputId: "input", references: [ref("ignored")] }),
	];
	const hypotheses = [hypothesis("h", { references: [ref("e3"), ref("e4")] })];
	const index = buildCitationIndex(report(["e1", "e2"]), findings, hypotheses);

	expect([...index.entries()]).toEqual([
		["e1", 1],
		["e2", 2],
		["e3", 3],
		["e4", 4],
	]);

	const findingView = buildFindingViews(findings, index)[0];
	expect(findingView.citations.map((citation) => citation.number)).toEqual([2, 3]);
	const hypothesisView = buildHypothesisViews(hypotheses, index)[0];
	expect(hypothesisView.citations.map((citation) => citation.number)).toEqual([3, 4]);
});

test("duplicate references within an item are cited once", () => {
	const findings = [finding("a", { references: [ref("e1"), ref("e1")] })];
	const index = buildCitationIndex(undefined, findings, []);
	expect(buildFindingViews(findings, index)[0].citations).toEqual([{ id: "e1", number: 1 }]);
});

test("finding relations resolve to rendered numbers", () => {
	const findings = [
		finding("a"),
		finding("b", {
			findingReferences: [
				{ versionId: "a", relation: "supports" },
				{ versionId: "missing", relation: "supports" },
				{ versionId: "a", relation: "mentions" },
			],
		}),
	];
	expect(buildFindingViews(findings, new Map())[1].relations).toEqual([
		"Supports finding 1",
		"Related to finding 1",
	]);
});

test("flags map to row badges", () => {
	const view = buildFindingViews(
		[finding("a", { invalidatedByVersionIds: ["x"], provisional: true })],
		new Map()
	)[0];
	const row = findingRow(view);
	expect(row.muted).toBe(true);
	expect(row.badges.map((badge) => badge.label)).toEqual(["Invalidated", "Provisional"]);

	const hypothesisView = buildHypothesisViews([hypothesis("h", { status: "supported" })], new Map())[0];
	expect(hypothesisRow(hypothesisView).badges.map((badge) => badge.label)).toEqual(["Supported"]);
});
