import { createMutation } from "@tanstack/svelte-query";
import { v4 as uuidv4 } from "uuid";
import { SvelteSet } from "svelte/reactivity";
import {
	createSystemAnalysisRelationshipMutation,
	updateSystemAnalysisRelationshipMutation,
	type CreateSystemAnalysisRelationshipAttributes,
	type SystemAnalysisRelationship,
	type SystemAnalysisRelationshipAttributes,
	type SystemAnalysisRelationshipControlAction,
	type SystemAnalysisRelationshipControlActionAttributes,
	type SystemAnalysisRelationshipFeedbackSignal,
	type SystemAnalysisRelationshipFeedbackSignalAttributes,
	type UpdateSystemAnalysisRelationshipAttributes,
} from "$lib/api";
import { useIncidentAnalysis } from "../../analysisState.svelte";
import { Context } from "runed";

const compareControlActions = (a: SystemAnalysisRelationshipControlAction, b: SystemAnalysisRelationshipControlAction) => {
	if (a.id !== b.id) return false;
	return (a.attributes.controlId === b.attributes.controlId) && (a.attributes.description === b.attributes.description);
}

const compareFeedbackSignals = (a: SystemAnalysisRelationshipFeedbackSignal, b: SystemAnalysisRelationshipFeedbackSignal) => {
	if (a.id !== b.id) return false;
	return (a.attributes.signalId === b.attributes.signalId) && (a.attributes.description === b.attributes.description);
}

// TODO: support this
type RelationshipKind =
	| 'request'    // API/Service requests
	| 'data'       // Data flow
	| 'telemetry'  // Monitoring/metrics
	| 'control';   // Control actions

const createRelationshipAttributesState = () => {
	let originalAttributes = $state<SystemAnalysisRelationshipAttributes>();
	let sourceId = $state<SystemAnalysisRelationshipAttributes["sourceId"]>("");
	let targetId = $state<SystemAnalysisRelationshipAttributes["targetId"]>("");
	let description = $state<SystemAnalysisRelationshipAttributes["description"]>("");
	let controlActions = $state<SystemAnalysisRelationshipAttributes["controlActions"]>([]);
	let feedbackSignals = $state<SystemAnalysisRelationshipAttributes["feedbackSignals"]>([]);

	let valid = $state(false);

	const descriptionChanged = $derived(originalAttributes?.description !== description);
	const controlsChanged = $derived.by(() => {
		const ogControls = originalAttributes?.controlActions ?? [];
		if (controlActions.length !== ogControls.length) return true;
		return controlActions.some((a, i) => !compareControlActions(ogControls[i], a))
	});
	const signalsChanged = $derived.by(() => {
		const ogSignals = originalAttributes?.feedbackSignals ?? [];
		if (feedbackSignals.length !== ogSignals.length) return true;
		return feedbackSignals.some((s, i) => !compareFeedbackSignals(ogSignals[i], s))
	});

	const initFrom = (a: SystemAnalysisRelationshipAttributes) => {
		originalAttributes = $state.snapshot(a);
		sourceId = $state.snapshot(a.sourceId);
		targetId = $state.snapshot(a.targetId);
		description = $state.snapshot(a.description);
		controlActions = $state.snapshot(a.controlActions);
		feedbackSignals = $state.snapshot(a.feedbackSignals);

		valid = true;
	}

	const initNew = (sourceId: string, targetId: string) => {
		initFrom({
			sourceId,
			targetId,
			description: "",
			controlActions: [],
			feedbackSignals: [],
		});
	}

	const onUpdate = () => {
		// TODO: check if attributes valid;
		valid = !!sourceId && !!targetId;
	}

	const includeControlAction = (controlId: string) => {
		updateControlAction({ controlId, description: "" });
	}

	const updateControlAction = (a: SystemAnalysisRelationshipControlActionAttributes) => {
		const idx = controlActions.findIndex(v => v.attributes.controlId === a.controlId);
		if (idx >= 0) { controlActions[idx].attributes = a }
		else { controlActions.push({ id: uuidv4(), attributes: a }) }
		onUpdate();
	}

	const removeControlAction = (id: string) => {
		controlActions = controlActions.filter(a => a.id !== id);
		onUpdate();
	}

	const includeFeedbackSignal = (signalId: string) => {
		updateFeedbackSignal({ signalId, description: "" });
	}

	const updateFeedbackSignal = (a: SystemAnalysisRelationshipFeedbackSignalAttributes) => {
		const idx = feedbackSignals.findIndex(v => v.attributes.signalId === a.signalId);
		if (idx >= 0) { feedbackSignals[idx].attributes = a }
		else { feedbackSignals.push({ id: uuidv4(), attributes: a }) }
		onUpdate();
	}

	const removeFeedbackSignal = (id: string) => {
		feedbackSignals = feedbackSignals.filter(a => a.id !== id);
		onUpdate();
	}

	return {
		initNew,
		initFrom,
		get targetId() { return targetId },
		get sourceId() { return sourceId },
		get description() { return description },
		set description(d: string) { description = d; onUpdate(); },
		get controlActions() { return controlActions },
		includeControlAction,
		updateControlAction,
		removeControlAction,
		get feedbackSignals() { return feedbackSignals },
		includeFeedbackSignal,
		updateFeedbackSignal,
		removeFeedbackSignal,
		snapshot() {
			return $state.snapshot({ sourceId, targetId, description, controlActions, feedbackSignals })
		},
		get valid() { return valid },
		get changed() { return descriptionChanged || controlsChanged || signalsChanged },
	}
}

export const relationshipAttributes = createRelationshipAttributesState();

type RelationshipDialogView = "closed" | "create" | "edit";

export class RelationshipDialogState {
	analysis = useIncidentAnalysis();
	
	view = $state<RelationshipDialogView>("closed");
	relationshipId = $state<string>();

	saveEnabled = $derived(relationshipAttributes.valid && (this.view === "create" || relationshipAttributes.changed));

	setCreating(sourceId: string, targetId: string) {
		this.view = "create";
		this.relationshipId = undefined;
		relationshipAttributes.initNew(sourceId, targetId);
	}

	setEditing(rel: SystemAnalysisRelationship) {
		this.view = "edit";
		this.relationshipId = rel.id;
		relationshipAttributes.initFrom(rel.attributes);
	};

	clear() {
		this.view = "closed";
		relationshipAttributes.initNew("", "");
		this.relationshipId = undefined;
	};

	onSuccess() {
		this.clear();
		alert("invalidate analysis query data");
		// this.analysis.invalidateQueryData();
	}

	createRelationshipMut = createMutation(() => ({ ...createSystemAnalysisRelationshipMutation(), onSuccess: this.onSuccess }))
	updateRelationshipMut = createMutation(() => ({ ...updateSystemAnalysisRelationshipMutation(), onSuccess: this.onSuccess }))

	loading = $derived(this.createRelationshipMut.isPending || this.updateRelationshipMut.isPending);

	doCreate() {
		const analysisId = $state.snapshot(this.analysis.analysisId);
		if (!analysisId) return;
		const attr = relationshipAttributes.snapshot();
		const attributes: CreateSystemAnalysisRelationshipAttributes = {
			sourceId: attr.sourceId,
			targetId: attr.targetId,
			description: attr.description,
			controlActions: attr.controlActions.map(a => a.attributes),
			feedbackSignals: attr.feedbackSignals.map(a => a.attributes),
		};
		this.createRelationshipMut.mutate({ path: { id: analysisId }, body: { attributes } });
	}

	doEdit() {
		if (!this.relationshipId) return;
		const attr = relationshipAttributes.snapshot();
		const attributes: UpdateSystemAnalysisRelationshipAttributes = {
			description: attr.description,
			controlActions: attr.controlActions.map(a => a.attributes),
			feedbackSignals: attr.feedbackSignals.map(s => s.attributes),
		};
		this.updateRelationshipMut.mutate({ path: { id: $state.snapshot(this.relationshipId) }, body: { attributes } });
	}

	onConfirm() {
		if (!relationshipAttributes.valid) return;
		if (this.view == "create") {
			this.doCreate();
		} else if (this.view == "edit") {
			this.doEdit();
		} else {
			console.error("Invalid view state");
		}
	};
};

const relationshipDialogCtx = new Context<RelationshipDialogState>("relationshipDialogState");
export const setRelationshipDialog = (r: RelationshipDialogState) => relationshipDialogCtx.set(r);
export const useRelationshipDialog = () => relationshipDialogCtx.get();;

export type RelationshipTrait = {
	id: string;
	attributes: {
		label: string;
		description: string;
	};
}

const createRelationshipTraitsState = () => {
	const includedSignalIds = $derived(
		new SvelteSet(relationshipAttributes.feedbackSignals.map((s) => s.attributes.signalId))
	);
	const includedControlIds = $derived(
		new SvelteSet(relationshipAttributes.controlActions.map((a) => a.attributes.controlId))
	);

	return {
		get includedSignalIds() { return includedSignalIds },
		get includedControlIds() { return includedControlIds },
	}
}

export const relationshipTraits = createRelationshipTraitsState();