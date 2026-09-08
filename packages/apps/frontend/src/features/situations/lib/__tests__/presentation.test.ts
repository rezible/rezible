import { describe, expect, test } from "bun:test";
import type { Situation } from "$lib/api";
import {
	attentionReasonLabels,
	byAttentionPriority,
	deriveSituationPresentation,
} from "../presentation";

const episode = (overrides: Partial<Parameters<typeof makeEpisode>[0]> = {}) => makeEpisode(overrides);

const makeEpisode = (attrs: {
	status?: "open" | "closed";
	lastObservedAt?: string;
	startedAt?: string;
}) => ({
	id: "ep-" + Math.random(),
	attributes: {
		status: attrs.status ?? "open",
		startedAt: attrs.startedAt ?? "2026-09-07T00:00:00Z",
		lastObservedAt: attrs.lastObservedAt ?? "2026-09-07T01:00:00Z",
	},
});

const makeSituation = (overrides: {
	status?: "open" | "closed";
	episodes?: ReturnType<typeof makeEpisode>[];
	investigation?: Situation["attributes"]["investigation"];
}) =>
	({
		id: "sit-1",
		attributes: {
			title: "Checkout degradation",
			summary: "",
			status: overrides.status ?? "open",
			evidenceRevision: 1,
			knowledgeEntityId: "ke-1",
			alertEpisodes: overrides.episodes ?? [],
			investigation: overrides.investigation,
			openedAt: "2026-09-07T00:00:00Z",
			updatedAt: "2026-09-07T00:00:00Z",
		},
	}) as unknown as Situation;

describe("deriveSituationPresentation", () => {
	test("active episodes mean active signals, never an assessment of safety", () => {
		const situation = makeSituation({
			episodes: [episode({ status: "open" })],
		});
		const presentation = deriveSituationPresentation(situation);
		expect(presentation.signalState).toBe("active");
		expect(presentation.openSignalCount).toBe(1);
		expect(presentation.attentionReasons).toContain("no-investigation");
	});

	test("quiet signals on an open situation stay visible as unassessed, not resolved", () => {
		const situation = makeSituation({
			episodes: [episode({ status: "closed" })],
		});
		const presentation = deriveSituationPresentation(situation);
		expect(presentation.signalState).toBe("quiet");
		expect(presentation.attentionReasons).toContain("signals-quiet-unassessed");
	});

	test("a completed report at the current evidence revision removes the unassessed reason", () => {
		const situation = makeSituation({
			episodes: [episode({ status: "closed" })],
			investigation: {
				id: "inv-1",
				attributes: {
					requestedRevision: 1,
					completedRevision: 1,
					evidenceRevision: 1,
					updatedAt: "2026-09-07T02:00:00Z",
					report: {
						text: "Investigated.",
						limitations: ["No deployment data"],
					},
				},
			} as Situation["attributes"]["investigation"],
		});
		const presentation = deriveSituationPresentation(situation);
		expect(presentation.signalState).toBe("quiet");
		expect(presentation.hasReport).toBe(true);
		expect(presentation.reportFresh).toBe(true);
		expect(presentation.attentionReasons).not.toContain("signals-quiet-unassessed");
		expect(presentation.reportLimitations).toEqual(["No deployment data"]);
	});

	test("a report older than the current evidence revision is stale and prioritized", () => {
		const situation = makeSituation({
			investigation: {
				id: "inv-1",
				attributes: {
					requestedRevision: 1,
					completedRevision: 1,
					evidenceRevision: 3,
					updatedAt: "2026-09-07T02:00:00Z",
				},
			} as Situation["attributes"]["investigation"],
		});
		const presentation = deriveSituationPresentation(situation);
		expect(presentation.reportStale).toBe(true);
		expect(presentation.attentionPriority).toBe(3);
	});

	test("a requested but incomplete investigation is running, not stale", () => {
		const situation = makeSituation({
			investigation: {
				id: "inv-1",
				attributes: {
					requestedRevision: 2,
					completedRevision: 1,
					evidenceRevision: 2,
					updatedAt: "2026-09-07T02:00:00Z",
				},
			} as Situation["attributes"]["investigation"],
		});
		const presentation = deriveSituationPresentation(situation);
		expect(presentation.investigationRunning).toBe(true);
		expect(presentation.reportStale).toBe(false);
		expect(presentation.attentionReasons).toContain("investigation-running");
	});

	test("closed situations never request operator attention", () => {
		const situation = makeSituation({ status: "closed" });
		const presentation = deriveSituationPresentation(situation);
		expect(presentation.attentionReasons).toEqual([]);
		expect(presentation.attentionPriority).toBe(0);
	});
});

describe("byAttentionPriority", () => {
	test("sorts stale-report situations before uninvestigated ones", () => {
		const staleReport = makeSituation({
			investigation: {
				id: "inv-1",
				attributes: {
					requestedRevision: 1,
					completedRevision: 1,
					evidenceRevision: 2,
					updatedAt: "2026-09-07T02:00:00Z",
				},
			} as Situation["attributes"]["investigation"],
		});
		const noInvestigation = makeSituation({});
		const sorted = [noInvestigation, staleReport].sort(byAttentionPriority);
		expect(sorted[0].id).toBe(staleReport.id);
	});

	test("every attention reason has a label", () => {
		for (const reason of ["report-stale", "no-investigation", "signals-quiet-unassessed", "investigation-running"] as const) {
			expect(attentionReasonLabels[reason]).toBeTruthy();
		}
	});
});
