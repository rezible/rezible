import type { GraphSubset } from "../graph";
import { mapModel } from "../model";

/** The part of GET /knowledge/model the map reads, as the backend serves it. */
export const testModel = mapModel({
	hierarchies: [
		{ value: "logical", categories: ["system_function", "system", "container", "component"] },
		{ value: "deployment", categories: ["infrastructure"] },
		{ value: "code", categories: ["code"] },
		{ value: "organization", categories: ["actor"] },
	],
	categories: [
		{ value: "system_function", hierarchy: "logical", level: 0 },
		{ value: "system", hierarchy: "logical", level: 1 },
		{ value: "container", hierarchy: "logical", level: 2 },
		{ value: "component", hierarchy: "logical", level: 3 },
		{ value: "infrastructure", hierarchy: "deployment" },
		{ value: "code", hierarchy: "code" },
		{ value: "actor", hierarchy: "organization" },
	],
	predicates: [
		{ value: "contains", family: "structure", rollsUp: false },
		{ value: "runs_on", family: "placement", rollsUp: false },
		{ value: "built_from", family: "implementation", rollsUp: false },
		{ value: "owns", family: "ownership", rollsUp: false },
		{ value: "uses", family: "flow", rollsUp: true },
		{ value: "calls", family: "flow", rollsUp: true },
		{ value: "reads_from", family: "flow", rollsUp: true },
		{ value: "writes_to", family: "flow", rollsUp: true },
		{ value: "publishes_to", family: "flow", rollsUp: true },
		{ value: "depends_on", family: "flow", rollsUp: true },
	],
});

/** A fixture graph with the test model. */
export const withTestModel = (graph: Omit<GraphSubset, "model">): GraphSubset => ({
	...graph,
	model: testModel,
});
