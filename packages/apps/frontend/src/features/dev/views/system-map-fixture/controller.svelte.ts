import { Context } from "runed";

import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { Point } from "$features/system/lib/system-map/geometry";
import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
import { reconcileSelection } from "$features/system/lib/system-map/selection";
import { systemMapFixtureScenarios, type FixtureScenarioId } from "./fixtures";

export class SystemMapFixtureController {
	scenario = $state<FixtureScenarioId>("shared");
	selection = $state.raw<MapSelection>();
	positions = $state.raw<Readonly<Partial<Record<string, Point>>>>({});
	dragEnabled = $state(true);

	graph = $derived<GraphSubset>(systemMapFixtureScenarios[this.scenario].graph);
	highlights = $derived<MapHighlights>({
		entityIds: this.selection?.kind === "entity" ? [this.selection.entityId] : [],
		relationshipIds:
			this.selection?.kind === "relationship"
				? [this.selection.relationshipId]
				: this.selection?.kind === "summary"
					? this.selection.relationshipIds
					: [],
	});

	sourceEntityCount = $derived(this.graph.entities.length);
	sourceRelationshipCount = $derived(
		this.graph.relationships.length + this.graph.unresolvedRelationships.length
	);
	unresolvedRelationshipCount = $derived(this.graph.unresolvedRelationships.length);
	partial = $derived(
		this.graph.enumeration.stopReason !== "exhausted" || this.unresolvedRelationshipCount > 0
	);

	setScenario = (value?: string) => {
		if (!value || !(value in systemMapFixtureScenarios) || value === this.scenario) return;
		this.scenario = value as FixtureScenarioId;
		this.positions = systemMapFixtureScenarios[this.scenario].positions ?? {};
		this.selection = reconcileSelection(systemMapFixtureScenarios[this.scenario].graph, this.selection);
	};

	setDragEnabled = (enabled: boolean) => {
		this.dragEnabled = enabled;
	};

	moveNode = (entityId: string, position: Point) => {
		this.positions = { ...this.positions, [entityId]: position };
	};

	setSelection = (selection: MapSelection | undefined) => {
		this.selection = selection;
	};
}

const ctx = new Context<SystemMapFixtureController>("SystemMapFixtureController");
export const initSystemMapFixtureController = () => ctx.set(new SystemMapFixtureController());
export const useSystemMapFixtureController = () => ctx.get();
