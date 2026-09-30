import { expect, test } from "bun:test";
import { incidentSlugUrl } from "./incident-url";

test("replaces only the incident path while preserving the view, query and hash", () => {
	const id = "22f06a48-a07b-4326-9839-4eb387a6704a";
	const url = new URL(`https://app.example/incidents/${id}/analysis?entry=${id}&kind=finding#details`);
	expect(incidentSlugUrl(url, id, "INC-42")).toBe(
		`/incidents/INC-42/analysis?entry=${id}&kind=finding#details`
	);
	expect(url.pathname).toBe(`/incidents/${id}/analysis`);
});
