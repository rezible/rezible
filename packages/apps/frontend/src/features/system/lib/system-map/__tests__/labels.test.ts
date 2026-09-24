import { describe, expect, test } from "bun:test";
import { nodePresentationForEntity } from "$features/system/components/system-map/map-node/presentation";

describe("system map labels", () => {
	test("shows the category and kind once each on nodes", () => {
		expect(
			nodePresentationForEntity({ id: "entity-42", category: "system_function", kind: "payments" })
		).toMatchObject({
			label: "payments",
			categoryLabel: "Function",
		});
		expect(
			nodePresentationForEntity({ id: "unknown-42", category: "future_domain", kind: "custom" })
		).toMatchObject({ label: "custom", categoryLabel: "future_domain" });
		expect(
			nodePresentationForEntity({ id: "code-42", category: "code", kind: "worker.ts" })
		).toMatchObject({
			label: "worker.ts",
			categoryLabel: "Code",
		});
	});
});
