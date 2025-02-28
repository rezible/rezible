import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { watch } from "runed";
import {
	getSystemAnalysisOptions,
	type SystemAnalysis,
	type SystemAnalysisRelationship,
	type SystemComponent,
} from "$lib/api";
import type { XYPosition } from "@xyflow/svelte";

const createAnalysisState = () => {
	let analysisId = $state<string>();
	let data = $state<SystemAnalysis>();
	let relationshipDialogOpen = $state(false);
	let editingRelationship = $state<SystemAnalysisRelationship>();

	const setup = (id: string) => {
		console.log("analysis setup", id);

		const queryClient = useQueryClient();

		const analysisQuery = createQuery(
			() => getSystemAnalysisOptions({ path: { id } }),
			queryClient
		);

		watch(() => analysisQuery.data, res => { if (res?.data) data = res.data });
	};

	const addComponent = async (component: SystemComponent, pos: XYPosition) => {
		console.log("add component");
	}

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
		addComponent,
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
