import { createMutation } from "@tanstack/svelte-query";
import {
	createSystemAnalysisRelationshipMutation,
	updateSystemAnalysisRelationshipMutation,
	type SystemAnalysisRelationship,
} from "$lib/api";

type RelationshipDialogView = "closed" | "create" | "edit";

const createRelationshipDialogState = () => {
	let view = $state<RelationshipDialogView>("closed");
	let editingRelationship = $state<SystemAnalysisRelationship>();
	let stateValid = $state(false);

	const setView = (v: RelationshipDialogView) => {
		view = v;
	}

	const clear = () => {
		setView("closed");
	};

	const goBack = () => {
		clear();
	}

	const createUpdateMutation = () => createMutation(() => ({ ...updateSystemAnalysisRelationshipMutation(), onSuccess: clear }));
	const createCreateMutation = () => createMutation(() => ({
		...createSystemAnalysisRelationshipMutation(), 
		
	}));

	let updateMut = $state<ReturnType<typeof createUpdateMutation>>();
	let createMut = $state<ReturnType<typeof createCreateMutation>>();

	const loading = $derived(updateMut?.isPending || createMut?.isPending);

	const setup = () => {
		updateMut = createUpdateMutation();
		createMut = createCreateMutation();
	};

	const setCreating = () => {
		setView("create");
	}

	const setEditing = (r: SystemAnalysisRelationship) => {
		setView("edit");
		editingRelationship = r;
	};

	const confirm = () => {
		
	};

	return {
		setup,
		get view() {
			return view;
		},
		get open() {
			return view !== "closed";
		},
		setCreating,
		setEditing,
		get editingRelationship() {
			return editingRelationship;
		},
		get stateValid() {
			return stateValid;
		},
		clear,
		goBack,
		confirm,
		get loading() {
			return loading;
		},
	};
};

export const relationshipDialog = createRelationshipDialogState();