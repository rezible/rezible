import { createMutation } from "@tanstack/svelte-query";
import { v4 as uuidv4 } from "uuid";
import {
	createSystemAnalysisRelationshipMutation,
	updateSystemAnalysisRelationshipMutation,
	type SystemAnalysisRelationship,
	type SystemAnalysisRelationshipAttributes,
	type SystemAnalysisRelationshipControlAction,
	type SystemAnalysisRelationshipControlActionAttributes,
	type SystemAnalysisRelationshipFeedbackSignal,
	type SystemAnalysisRelationshipFeedbackSignalAttributes,
} from "$lib/api";

const compareControlActions = (a: SystemAnalysisRelationshipControlAction, b: SystemAnalysisRelationshipControlAction) => {
	if (a.id !== b.id) return false;
	return (a.attributes.control_id === b.attributes.control_id) && (a.attributes.description === b.attributes.description);
}

const compareFeedbackSignals = (a: SystemAnalysisRelationshipFeedbackSignal, b: SystemAnalysisRelationshipFeedbackSignal) => {
	if (a.id !== b.id) return false;
	return (a.attributes.signal_id === b.attributes.signal_id) && (a.attributes.description === b.attributes.description);
}

// TODO: support this
type RelationshipKind =
  | 'request'    // API/Service requests
  | 'data'       // Data flow
  | 'telemetry'  // Monitoring/metrics
  | 'control';   // Control actions

const createRelationshipAttributesState = () => {
	let originalAttributes = $state<SystemAnalysisRelationshipAttributes>();
	let sourceId = $state<SystemAnalysisRelationshipAttributes["source_id"]>("");
	let targetId = $state<SystemAnalysisRelationshipAttributes["target_id"]>("");
	let description = $state<SystemAnalysisRelationshipAttributes["description"]>("");
	let controlActions = $state<SystemAnalysisRelationshipAttributes["control_actions"]>([]);
	let feedbackSignals = $state<SystemAnalysisRelationshipAttributes["feedback_signals"]>([]);
	
	let valid = $state(false);

	const descriptionChanged = $derived(originalAttributes?.description !== description);
	const controlsChanged = $derived.by(() => {
		const ogControls = originalAttributes?.control_actions ?? [];
		if (controlActions.length !== ogControls.length) return true;
		return controlActions.some((a, i) => !compareControlActions(ogControls[i], a))
	});
	const signalsChanged = $derived.by(() => {
		const ogSignals = originalAttributes?.feedback_signals ?? [];
		if (feedbackSignals.length !== ogSignals.length) return true;
		return feedbackSignals.some((s, i) => !compareFeedbackSignals(ogSignals[i], s))
	});

	const initFrom = (a: SystemAnalysisRelationshipAttributes) => {
		originalAttributes = $state.snapshot(a);
		sourceId = $state.snapshot(a.source_id);
		targetId = $state.snapshot(a.target_id);
		description = $state.snapshot(a.description);
		controlActions = $state.snapshot(a.control_actions);
		feedbackSignals = $state.snapshot(a.feedback_signals);

		valid = true;
	}

	const initNew = (sourceId: string, targetId: string) => {
		initFrom({
			source_id: $state.snapshot(sourceId),
			target_id: $state.snapshot(targetId),
			description: "",
			control_actions: [],
			feedback_signals: [],
		});
	}

	const onUpdate = () => {
		// TODO: check if attributes valid;
		valid = !!sourceId && !!targetId;
	}

	const setControlAction = (a: SystemAnalysisRelationshipControlActionAttributes) => {
		const idx = controlActions.findIndex(v => v.attributes.control_id === a.control_id);
		if (idx >= 0) { controlActions[idx].attributes = a }
		else { controlActions.push({id: uuidv4(), attributes: a}) }
		onUpdate();
	}

	const removeControlAction = (id: string) => {
		controlActions = controlActions.filter(a => a.id !== id);
		onUpdate();
	}

	const setFeedbackSignal = (a: SystemAnalysisRelationshipFeedbackSignalAttributes) => {
		const idx = feedbackSignals.findIndex(v => v.attributes.signal_id === a.signal_id);
		if (idx >= 0) { feedbackSignals[idx].attributes = a }
		else { feedbackSignals.push({id: uuidv4(), attributes: a}) }
		onUpdate();
	}

	const removeFeedbackSignal = (id: string) => {
		feedbackSignals = feedbackSignals.filter(a => a.id !== id);
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
		setControlAction,
		removeControlAction,
		get feedbackSignals() { return feedbackSignals },
		setFeedbackSignal,
		removeFeedbackSignal,
		snapshot() {
			return {
				source_id: $state.snapshot(sourceId),
				target_id: $state.snapshot(targetId),
				description: $state.snapshot(description),
				control_actions: $state.snapshot(controlActions),
				feedback_signals: $state.snapshot(feedbackSignals),
			}
		},
		get valid() { return valid },
		get changed() { return descriptionChanged || controlsChanged || signalsChanged },
	}
}

type RelationshipDialogView = "closed" | "create" | "edit";

const createRelationshipDialogState = () => {
	let view = $state<RelationshipDialogView>("closed");
	let relationshipId = $state<string>();
	let relationshipAttributes = createRelationshipAttributesState();

	const setView = (v: RelationshipDialogView) => {view = v}

	const clear = () => {
		setView("closed");
		relationshipAttributes.initNew("", "");
		relationshipId = undefined;
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
		relationshipId = undefined;
		relationshipAttributes.initNew(sourceId, targetId);
	}

	const setEditing = (rel: SystemAnalysisRelationship) => {
		setView("edit");
		relationshipId = rel.id;
		relationshipAttributes.initFrom(rel.attributes);
	};

	const confirm = () => {
		
		clear();
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
		get saveEnabled() {
			return relationshipAttributes.valid && (view === "create" || relationshipAttributes.changed);
		},
		clear,
		confirm,
		get loading() {
			return loading;
		},
	};
};

export const relationshipDialog = createRelationshipDialogState();