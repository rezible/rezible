import type { GraphSubset } from "$features/systems/lib/system-map/graph";
import type { MapDisplayOptions } from "$features/systems/lib/system-map/presentation";
import { Context } from "runed";
import { systemMapFixtureScenarios, type FixtureScenarioId } from "./fixtures";

const coverageLabel = (coverage: GraphSubset["coverage"]["parentMembership"]): string =>
	coverage.charAt(0).toUpperCase() + coverage.slice(1);

export class SystemMapFixtureController {
	scenario = $state<FixtureScenarioId>("hierarchy");
	showActors = $state(false);
	showAnnotations = $state(false);

	graph = $derived<GraphSubset>(systemMapFixtureScenarios[this.scenario].source);
	displayOptions = $derived<MapDisplayOptions>({
		showActors: this.showActors,
		showAnnotations: this.showAnnotations,
	});

	sourceEntityCount = $derived(this.graph.entities.length);
	sourceRelationshipCount = $derived(this.graph.relationships.length);

	parentMembershipCoverage = $derived(coverageLabel(this.graph.coverage.parentMembership));
	relationshipCoverage = $derived(coverageLabel(this.graph.coverage.relationships));

	setScenario = (value?: string) => {
		if (!value || !(value in systemMapFixtureScenarios) || value === this.scenario) return;

		this.scenario = value as FixtureScenarioId;
	};

	setShowActors = (showActors: boolean) => {
		this.showActors = showActors;
	};

	setShowAnnotations = (showAnnotations: boolean) => {
		this.showAnnotations = showAnnotations;
	};

}

const ctx = new Context<SystemMapFixtureController>("SystemMapFixtureController");
export const initSystemMapFixtureController = () => ctx.set(new SystemMapFixtureController());
export const useSystemMapFixtureController = () => ctx.get();