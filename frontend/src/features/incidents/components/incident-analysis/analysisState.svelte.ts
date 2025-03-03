import { createMutation, createQuery } from "@tanstack/svelte-query";
import {
	addSystemAnalysisComponentMutation,
	getSystemAnalysisOptions,
	type AddSystemAnalysisComponentData,
	type SystemAnalysisRelationship,
	type SystemComponent,
} from "$lib/api";
import type { XYPosition } from "@xyflow/svelte";

const createAnalysisState = () => {
	let analysisId = $state<string>();

	let relationshipDialogOpen = $state(false);
	let editingRelationship = $state<SystemAnalysisRelationship>();

	const makeAnalysisQuery = (id: string) => createQuery(() => getSystemAnalysisOptions({ path: { id } }));
	let analysisQuery = $state<ReturnType<typeof makeAnalysisQuery>>();

	const analysisData = $derived(analysisQuery?.data?.data);

	const makeAddComponentMutation = () => createMutation(() => addSystemAnalysisComponentMutation());
	let addComponentMut = $state<ReturnType<typeof makeAddComponentMutation>>();

	// const components = $derived(analysisData?.attributes.components ?? []);
	// const relationships = $derived(analysisData?.attributes.relationships ?? []);

	const setup = (id: string) => {
		analysisId = id;

		analysisQuery = makeAnalysisQuery(id);
		addComponentMut = makeAddComponentMutation();
	};

	const addComponent = async (component: SystemComponent, pos: XYPosition) => {
		if (!addComponentMut || !analysisId) return false;

		const path = { id: analysisId };
		const body = { attributes: { componentId: component.id, position: pos } };

		try {
			const resp = await addComponentMut.mutateAsync({ path, body});
			return resp.data;
		} catch (e) {
			return false;
		}
	}

	const setRelationshipDialogOpen = (open: boolean, editRel?: SystemAnalysisRelationship) => {
		relationshipDialogOpen = open;
		editingRelationship = editRel;
	};

	return {
		setup,
		get id() { return analysisId },
		get data() {
			return analysisData;
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
