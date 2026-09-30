import { expect, test } from "bun:test";
import type { SystemAnalysisEntry } from "$lib/api";
import { entryInput } from "./entry-input";

test("narrative edits preserve all subject kinds, including unavailable references", () => {
	const entry: SystemAnalysisEntry = {
		id: "entry",
		attributes: {
			analysisId: "analysis",
			kind: "finding",
			title: "Original",
			body: "Original body",
			sequence: 1,
			properties: {},
			version: 1,
			createdAt: "2026-09-30T00:00:00Z",
			updatedAt: "2026-09-30T00:00:00Z",
			subjects: [
				{ id: "one", attributes: { role: "affected", knowledgeEntityId: "entity", available: true } },
				{
					id: "two",
					attributes: { role: "context", knowledgeRelationshipId: "relationship", available: true },
				},
				{
					id: "three",
					attributes: { role: "supports", knowledgeEvidenceId: "evidence", available: false },
				},
			],
		},
	};
	const result = entryInput(
		{ title: "Edited", kind: "finding", body: "Revised body", occurredAt: null },
		entry
	);
	expect(JSON.parse(JSON.stringify(result))).toEqual({
		title: "Edited",
		kind: "finding",
		body: "Revised body",
		occurredAt: null,
		subjects: [
			{ role: "affected", knowledgeEntityId: "entity" },
			{ role: "context", knowledgeRelationshipId: "relationship" },
			{ role: "supports", knowledgeEvidenceId: "evidence" },
		],
	});
	expect(entry.attributes.title).toBe("Original");
});
