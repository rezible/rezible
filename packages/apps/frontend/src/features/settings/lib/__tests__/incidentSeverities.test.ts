import { expect, test } from "bun:test";

import type { IncidentSeverity } from "$lib/api";

import { sortSeveritiesByRank } from "../incidentSeverities";

const severity = (name: string, rank: number) =>
	({
		id: name,
		attributes: { name, rank, color: "", description: "", archived: false },
	}) as IncidentSeverity;

test("sorts severities by rank, then name", () => {
	const sorted = sortSeveritiesByRank([severity("SEV-3", 0), severity("Minor", 2), severity("SEV-1", 0)]);

	expect(sorted.map((item) => item.attributes.name)).toEqual(["SEV-1", "SEV-3", "Minor"]);
});
