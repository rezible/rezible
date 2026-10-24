import { describe, expect, it } from "bun:test";

import { responseActions } from "./response";

describe("responseActions", () => {
	it("offers mitigating or resolving a started incident", () => {
		expect(responseActions("started", true)).toEqual(["mitigated", "resolve"]);
	});

	it("offers a failed mitigation or resolving a mitigated incident", () => {
		expect(responseActions("mitigated", true)).toEqual(["mitigationFailed", "resolve"]);
	});

	it("offers nothing once resolved", () => {
		expect(responseActions("resolved", true)).toEqual([]);
	});

	it("offers nothing while incident management is off", () => {
		expect(responseActions("started", false)).toEqual([]);
		expect(responseActions("mitigated", false)).toEqual([]);
	});
});
