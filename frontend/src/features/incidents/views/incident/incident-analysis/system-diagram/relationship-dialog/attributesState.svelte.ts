import type { SystemAnalysisRelationshipAttributes, SystemComponentRelationshipAttributes } from "$lib/api";
import { SvelteSet } from "svelte/reactivity";
import { v4 as uuidv4 } from "uuid";

// const compareControlActions = (a: SystemRelationshipControlAction, b: SystemAnalysisRelationshipControlAction) => {
// 	if (a.id !== b.id) return false;
// 	return (a.attributes.controlId === b.attributes.controlId) && (a.attributes.description === b.attributes.description);
// }

// const compareFeedbackSignals = (a: SystemAnalysisRelationshipFeedbackSignal, b: SystemAnalysisRelationshipFeedbackSignal) => {
// 	if (a.id !== b.id) return false;
// 	return (a.attributes.signalId === b.attributes.signalId) && (a.attributes.description === b.attributes.description);
// }

const createRelationshipAttributesState = () => {
	let originalAttributes = $state<SystemAnalysisRelationshipAttributes>();
	let sourceId = $state<SystemComponentRelationshipAttributes["sourceId"]>("");
	let targetId = $state<SystemComponentRelationshipAttributes["targetId"]>("");
	let description = $state<SystemAnalysisRelationshipAttributes["description"]>("");
	// let controlActions = $state<SystemAnalysisRelationshipAttributes["controlActions"]>([]);
	// let feedbackSignals = $state<SystemAnalysisRelationshipAttributes["feedbackSignals"]>([]);

	let valid = $state(false);

	const descriptionChanged = $derived(originalAttributes?.description !== description);
	const controlsChanged = $derived.by(() => {
		// const ogControls = originalAttributes?.controlActions ?? [];
		// if (controlActions.length !== ogControls.length) return true;
		// return controlActions.some((a, i) => !compareControlActions(ogControls[i], a))
	});
	const signalsChanged = $derived.by(() => {
		// const ogSignals = originalAttributes?.feedbackSignals ?? [];
		// if (feedbackSignals.length !== ogSignals.length) return true;
		// return feedbackSignals.some((s, i) => !compareFeedbackSignals(ogSignals[i], s))
	});

	const initFrom = (a: SystemComponentRelationshipAttributes) => {
		// originalAttributes = $state.snapshot(a);
		// sourceId = $state.snapshot(a.sourceId);
		// targetId = $state.snapshot(a.targetId);
		// description = $state.snapshot(a.description);
		// controlActions = $state.snapshot(a.controlActions);
		// feedbackSignals = $state.snapshot(a.feedbackSignals);

		valid = true;
	}

	const initNew = (sourceId: string, targetId: string) => {
		initFrom({
			sourceId,
			targetId,
			description: "",
			// controlActions: [],
			// feedbackSignals: [],
		});
	}

	const onUpdate = () => {
		// TODO: check if attributes valid;
		valid = !!sourceId && !!targetId;
	}

	const includeControlAction = (controlId: string) => {
		updateControlAction({ controlId, description: "" });
	}

	const updateControlAction = (a: any) => {
	// const updateControlAction = (a: SystemAnalysisRelationshipControlActionAttributes) => {
		// const idx = controlActions.findIndex(v => v.attributes.controlId === a.controlId);
		// if (idx >= 0) { controlActions[idx].attributes = a }
		// else { controlActions.push({ id: uuidv4(), attributes: a }) }
		onUpdate();
	}

	const removeControlAction = (id: string) => {
		// controlActions = controlActions.filter(a => a.id !== id);
		onUpdate();
	}

	const includeFeedbackSignal = (signalId: string) => {
		updateFeedbackSignal({ signalId, description: "" });
	}

	// const updateFeedbackSignal = (a: SystemAnalysisRelationshipFeedbackSignalAttributes) => {
	const updateFeedbackSignal = (a: any) => {
		// const idx = feedbackSignals.findIndex(v => v.attributes.signalId === a.signalId);
		// if (idx >= 0) { feedbackSignals[idx].attributes = a }
		// else { feedbackSignals.push({ id: uuidv4(), attributes: a }) }
		onUpdate();
	}

	const removeFeedbackSignal = (id: string) => {
		// feedbackSignals = feedbackSignals.filter(a => a.id !== id);
		onUpdate();
	}

	return {
		initNew,
		initFrom,
		get targetId() { return targetId },
		get sourceId() { return sourceId },
		get description() { return description },
		set description(d: string) { description = d; onUpdate(); },
		get controlActions() { return [] },
		includeControlAction,
		updateControlAction,
		removeControlAction,
		get feedbackSignals() { return [] },
		includeFeedbackSignal,
		updateFeedbackSignal,
		removeFeedbackSignal,
		snapshot() {
			return $state.snapshot({ sourceId, targetId, description })
		},
		get valid() { return valid },
		get changed() { return descriptionChanged || controlsChanged || signalsChanged },
	}
}

export const relationshipAttributes = createRelationshipAttributesState();

export type RelationshipTrait = {
	id: string;
	attributes: {
		label: string;
		description: string;
	};
}

const createRelationshipTraitsState = () => {
	const includedSignalIds = $derived(
		new SvelteSet(/*relationshipAttributes.feedbackSignals.map((s) => s.attributes.signalId)*/)
	);
	const includedControlIds = $derived(
		new SvelteSet(/*relationshipAttributes.controlActions.map((a) => a.attributes.controlId)*/)
	);

	return {
		get includedSignalIds() { return includedSignalIds },
		get includedControlIds() { return includedControlIds },
	}
}

export const relationshipTraits = createRelationshipTraitsState();