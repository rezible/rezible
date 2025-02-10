import { createMutation } from "@tanstack/svelte-query";
import {
	createSystemAnalysisRelationshipMutation,
	updateSystemAnalysisRelationshipMutation,
	type SystemAnalysisRelationship,
	type SystemAnalysisRelationshipAttributes,
	type SystemAnalysisRelationshipControlAction,
	type SystemAnalysisRelationshipFeedbackSignal,
} from "$lib/api";

const createRelationshipAttributesState = () => {
	let sourceId = $state<SystemAnalysisRelationshipAttributes["source_id"]>("");
	let targetId = $state<SystemAnalysisRelationshipAttributes["target_id"]>("");
	let description = $state<SystemAnalysisRelationshipAttributes["description"]>("");
	let controlActions = $state<SystemAnalysisRelationshipAttributes["control_actions"]>([]);
	let feedbackSignals = $state<SystemAnalysisRelationshipAttributes["feedback_signals"]>([]);
	
	let valid = $state(false);

	const initFrom = (a: SystemAnalysisRelationshipAttributes) => {
		sourceId = a.source_id;
		targetId = a.target_id;
		description = a.description;
		controlActions = a.control_actions;
		feedbackSignals = a.feedback_signals;
		valid = true;
	}

	const initNew = (sourceId: string, targetId: string) => {
		initFrom({
			source_id: sourceId,
			target_id: targetId,
			description: "",
			control_actions: [],
			feedback_signals: [],
		});
	}

	const onUpdate = () => {
		// TODO: check if attributes valid;
		valid = !!sourceId && !!targetId;
	}

	const updateControlAction = (c: SystemAnalysisRelationshipControlAction) => {
		const idx = controlActions.findIndex(v => v.id === c.id);
		if (idx >= 0) { controlActions[idx] = c }
		else { controlActions.push(c) }
		onUpdate();
	}

	const updateFeedbackSignal = (s: SystemAnalysisRelationshipFeedbackSignal) => {
		const idx = feedbackSignals.findIndex(v => v.id === s.id);
		if (idx >= 0) { feedbackSignals[idx] = s }
		else { feedbackSignals.push(s) }
		onUpdate();
	}

	// this is gross but oh well
	return {
		initNew,
		initFrom,
		get targetId() { return targetId },
		get sourceId() { return sourceId },
		get description() { return description },
		set description(d: string) { description = d; onUpdate(); },
		get controlActions() { return controlActions },
		updateControlAction,
		get feedbackSignals() { return feedbackSignals },
		updateFeedbackSignal,
		asAttributes(): SystemAnalysisRelationshipAttributes {
			return {
				source_id: $state.snapshot(sourceId),
				target_id: $state.snapshot(targetId),
				description: $state.snapshot(description),
				control_actions: $state.snapshot(controlActions),
				feedback_signals: $state.snapshot(feedbackSignals),
			}
		},
		get valid() { return valid },
	}
}

type RelationshipDialogView = "closed" | "create" | "edit";

const createRelationshipDialogState = () => {
	let view = $state<RelationshipDialogView>("closed");
	let editingRelationship = $state<SystemAnalysisRelationship>();
	let relationshipAttributes = createRelationshipAttributesState();
	let stateValid = $state(false);

	const setView = (v: RelationshipDialogView) => {view = v}

	const clear = () => {
		setView("closed");
		relationshipAttributes = createRelationshipAttributesState();
		editingRelationship = undefined;
		stateValid = false;
	};

	const onSuccess = () => {

	}

	const makeUpdateMutation = () => createMutation(() => ({
		...updateSystemAnalysisRelationshipMutation(),
		onSuccess,
	}));
	const makeCreateMutation = () => createMutation(() => ({
		...createSystemAnalysisRelationshipMutation(), 
		onSuccess,
	}));

	let updateMut = $state<ReturnType<typeof makeUpdateMutation>>();
	let createMut = $state<ReturnType<typeof makeCreateMutation>>();

	const loading = $derived(updateMut?.isPending || createMut?.isPending);

	const setup = () => {
		updateMut = makeUpdateMutation();
		createMut = makeCreateMutation();
	};

	const setCreating = (sourceId: string, targetId: string) => {
		setView("create");
		relationshipAttributes.initNew(sourceId, targetId);
	}

	const setEditing = (rel: SystemAnalysisRelationship) => {
		setView("edit");
		editingRelationship = rel;
		relationshipAttributes.initFrom(rel.attributes);
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
		get attributes() {
			return relationshipAttributes;
		},
		get stateValid() {
			return stateValid;
		},
		clear,
		confirm,
		get loading() {
			return loading;
		},
	};
};

export const relationshipDialog = createRelationshipDialogState();