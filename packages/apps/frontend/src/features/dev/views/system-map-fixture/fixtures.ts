import type { GraphSubset } from "$features/system/lib/system-map/graph";
import { largeFixture } from "./large-fixture";
import { fullFixture } from "./terrain-fixture";

export type FixtureScenarioId = "terrain" | "large" | "empty";

/** A fixture graph; the page adds the model the backend serves. */
export type FixtureGraph = Omit<GraphSubset, "model">;

export type FixtureScenario = {
	graph: FixtureGraph;
};

const emptyGraph: FixtureGraph = {
	entities: [],
	relationships: [],
	unresolvedRelationships: [],
	enumeration: { scope: "fixture", stopReason: "exhausted" },
};

export const systemMapFixtureScenarios: Readonly<Record<FixtureScenarioId, FixtureScenario>> = {
	terrain: {
		graph: fullFixture,
	},
	large: {
		graph: largeFixture,
	},
	empty: {
		graph: emptyGraph,
	},
};
