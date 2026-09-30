import { expect, test } from "bun:test";
import type { ExpandableIncidentSeverityAttributes } from "$lib/api";
import { incidentResponseStatus, incidentSeverityStatus, retrospectiveStatus } from "./status";

test("incidentResponseStatus labels", () => {
	const cases = [
		["started", "Active", "danger"],
		["mitigated", "Mitigated", "warning"],
		["resolved", "Resolved", "success"],
		["unknown", "Unknown", "neutral"],
	] as const;
	for (const [state, label, tone] of cases) {
		const status = incidentResponseStatus(state);
		expect([status.label, status.tone]).toEqual([label, tone]);
		expect(status.description).toBe("Incident response state from the incident tool");
	}
});

test("retrospectiveStatus labels", () => {
	const cases = [
		[undefined, "Not started", "neutral"],
		["draft", "Retrospective", "info"],
		["in_review", "In review", "warning"],
		["meeting", "Review meeting", "warning"],
		["closed", "Closed", "success"],
	] as const;
	for (const [state, label, tone] of cases) {
		const status = retrospectiveStatus(state);
		expect([status.label, status.tone]).toEqual([label, tone]);
		expect(status.description).toBe("Rezible review stage");
	}
});

test("incidentSeverityStatus uses the severity rank", () => {
	const severity = (rank: number): ExpandableIncidentSeverityAttributes => ({
		id: `sev-${rank}`,
		attributes: { name: `SEV-${rank}`, rank, archived: false, color: "", description: "" },
	});

	expect(incidentSeverityStatus(undefined).label).toBe("No severity");
	expect(incidentSeverityStatus({ id: "missing" }).label).toBe("No severity");

	const cases = [
		[1, "danger"],
		[2, "warning"],
		[3, "neutral"],
	] as const;
	for (const [rank, tone] of cases) {
		const status = incidentSeverityStatus(severity(rank));
		expect([status.label, status.tone]).toEqual([`SEV-${rank}`, tone]);
	}
});
