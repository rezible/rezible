import { expect, test } from "bun:test";
import type { Incident } from "$lib/api";
import { incidentListRow } from "./row";

test("incident rows dedupe and limit service names", () => {
	const impact = (name: string | undefined) => ({
		id: name ?? "none",
		note: "",
		source: "",
		knowledgeEntity: {
			id: "entity",
			attributes: name ? { state: { displayName: name } } : undefined,
		},
	});
	const incident = {
		id: "incident",
		attributes: {
			reference: "INC-1",
			title: "Incident",
			summary: " Summary ",
			responseState: "started",
			openedAt: "2026-05-14T10:00:00Z",
			resolvedAt: null,
			updatedAt: "2026-05-14T11:00:00Z",
			impacts: ["a", "b", "a", undefined, "c", "d"].map(impact),
		},
	} as unknown as Incident;

	const row = incidentListRow(incident);
	expect(row.services).toEqual(["a", "b", "c"]);
	expect(row.moreServices).toBe(1);
	expect(row.summary).toBe("Summary");
	expect(row.response.label).toBe("Active");
	expect(row.severity).toBeUndefined();
});
