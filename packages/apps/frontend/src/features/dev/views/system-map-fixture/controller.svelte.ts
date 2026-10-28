import { Context } from "runed";
import { tick } from "svelte";

import type { SystemMapHandle } from "$features/system/components/system-map";
import type { GraphSubset } from "$features/system/lib/system-map/graph";
import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
import { reconcileSelection } from "$features/system/lib/system-map/selection";
import { systemMapFixtureScenarios, type FixtureScenarioId } from "./fixtures";

export class SystemMapFixtureController {
	scenario = $state<FixtureScenarioId>("terrain");
	selection = $state.raw<MapSelection>();
	mapView = $state<"structure" | "deployment">("structure");
	deploymentFocusTargetId = $state<string>();
	mapHandle = $state<SystemMapHandle>();

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
		this.selection = reconcileSelection(systemMapFixtureScenarios[this.scenario].graph, this.selection);
	};

	setSelection = (selection: MapSelection | undefined) => {
		this.selection = selection;
	};

	setMapView = (value?: string) => {
		if (value !== "structure" && value !== "deployment") return;
		this.deploymentFocusTargetId = undefined;
		this.mapView = value;
	};

	exploreDeployment = (targetId: string) => {
		this.deploymentFocusTargetId = targetId;
		this.mapView = "deployment";
	};

	showInStructure = async (entityId: string) => {
		this.setMapView("structure");
		// The hidden map has no size until the frame after it is shown measures it.
		await tick();
		await new Promise((resolve) => requestAnimationFrame(resolve));
		await new Promise((resolve) => requestAnimationFrame(resolve));
		this.mapHandle?.reveal({ kind: "entity", entityId });
	};
}

const ctx = new Context<SystemMapFixtureController>("SystemMapFixtureController");
export const initSystemMapFixtureController = () => ctx.set(new SystemMapFixtureController());
export const useSystemMapFixtureController = () => ctx.get();
