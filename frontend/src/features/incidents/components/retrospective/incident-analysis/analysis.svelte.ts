import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { watch } from "runed";
import {
	getSystemAnalysisOptions,
	type SystemAnalysis,
	type SystemAnalysisComponent,
	type SystemAnalysisRelationship,
	type SystemComponent,
} from "$lib/api";
import { incidentCtx } from "$features/incidents/lib/context";

const createAnalysisState = () => {
	let analysisId = $state<string>();
	let data = $state<SystemAnalysis>();
	let addingComponent = $state<SystemComponent>();
	let relationshipDialogOpen = $state(false);
	let editingRelationship = $state<SystemAnalysisRelationship>();

	const setup = () => {
		analysisId = incidentCtx.get().attributes.systemAnalysisId;

		const queryClient = useQueryClient();

		const analysisQuery = createQuery(
			() => ({
				...getSystemAnalysisOptions({ path: { id: analysisId ?? "" } }),
				enabled: !!analysisId,
			}),
			queryClient
		);

		watch(
			() => analysisQuery.data,
			(body) => {
				if (!body?.data) return;
				data = body.data;
			}
		);
	};

	const setAddingComponent = (c?: SystemComponent) => {
		addingComponent = c;
	};

	const setRelationshipDialogOpen = (open: boolean, editRel?: SystemAnalysisRelationship) => {
		relationshipDialogOpen = open;
		editingRelationship = editRel;
	};

	return {
		setup,
		get id() { return analysisId },
		get data() {
			return data;
		},
		setAddingComponent,
		get addingComponent() {
			return addingComponent;
		},
		get relationshipDialogOpen() {
			return relationshipDialogOpen;
		},
		setRelationshipDialogOpen,
		get editingRelationship() {
			return editingRelationship;
		},
	};
};

export const analysis = createAnalysisState();
