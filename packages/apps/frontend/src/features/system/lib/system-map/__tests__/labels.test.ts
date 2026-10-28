import { describe, expect, test } from "bun:test";
import { entityCategoryLabel, entityLabel } from "../labels";

describe("system map labels", () => {
	test("labels entities by display name and preserves unrecognized categories", () => {
		expect(
			entityLabel({ id: "entity-42", category: "container", kind: "service", name: "checkout-api" })
		).toBe("checkout-api");
		expect(entityCategoryLabel("container")).toBe("Container");
		expect(entityCategoryLabel("future_domain")).toBe("future_domain");
	});

	test("falls back to the kind when the entity is unnamed", () => {
		expect(entityLabel({ id: "code-42", category: "code", kind: "worker.ts", name: "" })).toBe(
			"worker.ts"
		);
	});
});
