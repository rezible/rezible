import { createMutation, createQuery, QueryClient, useQueryClient } from "@tanstack/svelte-query";
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

	let queryClient = $state<QueryClient>();
	const analysisQueryOpts = $derived(getSystemAnalysisOptions({ path: { id: (analysisId ?? "") } }))
	const makeAnalysisQuery = () => createQuery(() => ({...analysisQueryOpts, enabled: !!analysisId}));
	let analysisQuery = $state<ReturnType<typeof makeAnalysisQuery>>();

	const analysisData = $derived(analysisQuery?.data?.data);

	const invalidateQueryData = () => {
		queryClient?.invalidateQueries(analysisQueryOpts);
	}

	const makeAddComponentMutation = () => createMutation(() => ({...addSystemAnalysisComponentMutation(), onSuccess: invalidateQueryData}));
	let addComponentMut = $state<ReturnType<typeof makeAddComponentMutation>>();

	// const components = $derived(analysisData?.attributes.components ?? []);
	// const relationships = $derived(analysisData?.attributes.relationships ?? []);

	const setup = (id: string) => {
		analysisId = id;
		queryClient = useQueryClient();

		analysisQuery = makeAnalysisQuery();
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
		invalidateQueryData,
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
