import { describe, expect, test } from "bun:test";

import {
	connectionEndpointOpacity,
	connectionLabel,
	connectionLabelOpacity,
	connectionRouteOpacities,
} from "$features/system/components/system-map/map-connection/presentation";
import { nodePresentationForEntity } from "$features/system/components/system-map/map-node/presentation";
import type { MapConnection } from "../presentation";

describe("system map labels", () => {
	test("shows the category and kind once each on nodes", () => {
		expect(
			nodePresentationForEntity({ id: "entity-42", category: "system_function", kind: "payments" })
		).toEqual({
			label: "payments",
			categoryLabel: "Function",
		});
		expect(
			nodePresentationForEntity({ id: "unknown-42", category: "future_domain", kind: "custom" })
		).toEqual({ label: "custom", categoryLabel: "future_domain" });
		expect(nodePresentationForEntity({ id: "code-42", category: "code", kind: "worker.ts" })).toEqual({
			label: "worker.ts",
			categoryLabel: "Code",
		});
	});

	test("raises hovered edges, labels, and displayed endpoints to full opacity", () => {
		const variants = [0.18, 0.27];
		const hoveredOpacity = connectionRouteOpacities(variants, true, true);

		expect(hoveredOpacity[0] / hoveredOpacity[1]).toBeCloseTo(variants[0] / variants[1]);
		expect(hoveredOpacity.reduce((total, opacity) => total + opacity, 0)).toBe(1);
		expect(connectionRouteOpacities(variants, false, true)).toEqual([0.045, 0.0675]);
		expect(connectionLabelOpacity(0.18, true, true)).toBe(1);
		expect(connectionEndpointOpacity(0.18, true)).toBe(1);
		expect(connectionEndpointOpacity(0.18, false)).toBe(0.18);
	});

	test("uses relationship predicates as canvas connection labels", () => {
		const relationship: MapConnection = {
			id: "relationship-1",
			endpoints: ["a", "b"],
			predicate: "runs_on",
			classification: "direct",
			sourceRelationshipIds: ["relationship-1"],
		};
		const summary: MapConnection = {
			...relationship,
			id: "summary-1",
			classification: "summary",
			sourceRelationshipIds: ["relationship-1", "relationship-2"],
		};

		expect(connectionLabel(relationship)).toBe("runs on");
		expect(connectionLabel(summary)).toBe("runs on · 2");
	});
});
