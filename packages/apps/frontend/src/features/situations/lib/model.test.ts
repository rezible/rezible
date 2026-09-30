import { expect, test } from "bun:test";
import type { InvestigationReport } from "$lib/api";
import { reportSummary, situationConclusion } from "./model";

const report = (text: string, summary = ""): InvestigationReport => ({
	id: "report",
	attributes: {
		agentTurnId: "turn",
		createdAt: "2026-05-14T04:45:00Z",
		provisional: false,
		references: [],
		summary,
		text,
		turnStatus: "completed",
	},
});

test("reportSummary prefers the structured summary", () => {
	expect(reportSummary(report("# Title\n\nFirst paragraph.", " Stated conclusion. ").attributes)).toBe(
		"Stated conclusion."
	);
	expect(reportSummary(report("# Title\n\nFirst paragraph.").attributes)).toBe("First paragraph.");
	expect(reportSummary(report("# Title only").attributes)).toBeUndefined();
});

test("situationConclusion maps query states", () => {
	expect(situationConclusion(undefined)).toEqual({ kind: "none" });
	expect(situationConclusion({ isPending: true, isError: false })).toEqual({ kind: "loading" });
	expect(situationConclusion({ isPending: false, isError: true })).toEqual({ kind: "none" });
	expect(
		situationConclusion({ isPending: false, isError: false, data: { data: report("Paragraph.") } })
	).toEqual({
		kind: "text",
		text: "Paragraph.",
	});
	expect(
		situationConclusion({ isPending: false, isError: false, data: { data: report("# Heading") } })
	).toEqual({
		kind: "none",
	});
});
