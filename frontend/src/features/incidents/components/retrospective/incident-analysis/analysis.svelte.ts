import { incidentCtx } from "$src/features/incidents/lib/context";
import { getSystemAnalysisOptions, type SystemAnalysis, type SystemAnalysisComponent, type SystemAnalysisRelationship } from "$src/lib/api";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { watch } from "runed";

const createAnalysisState = () => {
	let analysisId = $state<string>();
	let data = $state<SystemAnalysis>();
	let editingComponent = $state<SystemAnalysisComponent>();
	let editingRelationship = $state<SystemAnalysisRelationship>();


	const setup = () => {
		analysisId = incidentCtx.get().attributes.system_analysis_id;
		
		const queryClient = useQueryClient();

		const analysisQuery = createQuery(() => ({
			...getSystemAnalysisOptions({path: {id: analysisId ?? ""}}),
			enabled: !!analysisId,
		}), queryClient);

		watch(() => analysisQuery.data, body => {
			if (!body?.data) return;
			data = body.data;
		});
	}

	const setEditingComponent = (c?: SystemAnalysisComponent) => {editingComponent = c};
	const setEditingRelationship = (r?: SystemAnalysisRelationship) => {editingRelationship = r};

	return {
		setup,
		get data() { return data },
		setEditingComponent,
		get editingComponent() { return editingComponent },
		setEditingRelationship,
		get editingRelationship() { return editingRelationship },
	}
}

export const analysis = createAnalysisState();