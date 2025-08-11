<script lang="ts" module>
	type RelationshipFeedbackSignal = {
		feedback: SystemAnalysisRelationshipFeedbackSignal;
		signal: SystemComponentSignal;
	};

	type RelationshipControlAction = {
		action: SystemAnalysisRelationshipControlAction;
		control: SystemComponentControl;
	};
</script>

<script lang="ts">
	import Button from "$components/button/Button.svelte";
	import type {
		SystemAnalysisRelationshipControlAction,
		SystemAnalysisRelationshipFeedbackSignal,
		SystemComponent,
		SystemComponentControl,
		SystemComponentSignal,
	} from "$lib/api";
	import { SvelteMap } from "svelte/reactivity";
	import { relationshipAttributes } from "./attributesState.svelte";
	import LabelDescriptionEditor from "./LabelDescriptionEditor.svelte";
	import { mdiMinus, mdiPencil } from "@mdi/js";
	import Header from "$components/header/Header.svelte";

	type Props = {
		source: SystemComponent;
		target: SystemComponent;
	};
	const { source, target }: Props = $props();

	const sourceSignals = $derived(new SvelteMap(source.attributes.signals.map((s) => [s.id, s])));
	const targetSignals = $derived(new SvelteMap(target.attributes.signals.map((s) => [s.id, s])));

	const sourceControls = $derived(new SvelteMap(source.attributes.controls.map((s) => [s.id, s])));
	const targetControls = $derived(new SvelteMap(target.attributes.controls.map((s) => [s.id, s])));

	const [sourceFeedbackSignals, targetFeedbackSignals] = $derived.by(() => {
		const src: RelationshipFeedbackSignal[] = [];
		const tgt: RelationshipFeedbackSignal[] = [];
		relationshipAttributes.feedbackSignals.forEach((feedback) => {
			const signalId = feedback.attributes.signalId;
			const sourceSignal = sourceSignals.get(signalId);
			if (sourceSignal) src.push({ feedback, signal: sourceSignal });
			const targetSignal = targetSignals.get(signalId);
			if (targetSignal) tgt.push({ feedback, signal: targetSignal });
		});
		return [src, tgt];
	});

	const [sourceControlActions, targetControlActions] = $derived.by(() => {
		const src: RelationshipControlAction[] = [];
		const tgt: RelationshipControlAction[] = [];
		relationshipAttributes.controlActions.forEach((action) => {
			const controlId = action.attributes.controlId;
			const sourceControl = sourceControls.get(controlId);
			if (sourceControl) src.push({ action, control: sourceControl });
			const targetControl = targetControls.get(controlId);
			if (targetControl) tgt.push({ action, control: targetControl });
		});
		return [src, tgt];
	});

	let editingAction = $state<SystemAnalysisRelationshipControlAction>();
	const updateEditingAction = () => {
		if (!editingAction) return;
		relationshipAttributes.updateControlAction($state.snapshot(editingAction.attributes));
		editingAction = undefined;
	};

	let editingFeedback = $state<SystemAnalysisRelationshipFeedbackSignal>();
	const updateEditingFeedback = () => {
		if (!editingFeedback) return;
		relationshipAttributes.updateFeedbackSignal($state.snapshot(editingFeedback.attributes));
		editingFeedback = undefined;
	};
</script>

<div class="flex flex-col gap-2">
	{@render componentFeedbackLoops(source.attributes.name, sourceFeedbackSignals, sourceControlActions)}
	{@render componentFeedbackLoops(target.attributes.name, targetFeedbackSignals, targetControlActions)}
</div>

{#snippet componentFeedbackLoops(
	title: string,
	feedbacks: RelationshipFeedbackSignal[],
	actions: RelationshipControlAction[]
)}
	<div class="border p-1">
		<Header {title} />

		<Header title={feedbacks.length > 0 ? "Sends Feedback" : "Supplies No Feedback"} />
		{#if feedbacks.length > 0}
			{#if editingFeedback}
				{@const signalId = editingFeedback.attributes.signalId}
				{@const ctrl = sourceSignals.get(signalId) || targetSignals.get(signalId)}
				<span>{ctrl?.attributes.label ?? "Editing"}</span>
				<LabelDescriptionEditor
					bind:description={editingFeedback.attributes.description}
					onCancel={() => {
						editingFeedback = undefined;
					}}
					onConfirm={updateEditingFeedback}
				/>
			{:else}
				{#each feedbacks as { feedback, signal }}
					{@render loopCard(
						signal.attributes.label,
						feedback.attributes.description,
						() => (editingFeedback = $state.snapshot(feedback)),
						() => relationshipAttributes.removeFeedbackSignal(feedback.id)
					)}
				{/each}
			{/if}
		{/if}

		<Header title={actions.length > 0 ? "Is Controlled via" : "Exposes No Controls"} />
		{#if actions.length > 0}
			{#if editingAction}
				{@const controlId = editingAction.attributes.controlId}
				{@const ctrl = sourceControls.get(controlId) || targetControls.get(controlId)}
				<span>{ctrl?.attributes.label ?? "Editing"}</span>
				<LabelDescriptionEditor
					bind:description={editingAction.attributes.description}
					onCancel={() => {
						editingAction = undefined;
					}}
					onConfirm={updateEditingAction}
				/>
			{:else}
				{#each actions as { control, action }}
					{@render loopCard(
						control.attributes.label,
						action.attributes.description,
						() => (editingAction = $state.snapshot(action)),
						() => relationshipAttributes.removeControlAction(action.id)
					)}
				{/each}
			{/if}
		{/if}
	</div>
{/snippet}

{#snippet loopCard(label: string, description: string, onEdit: VoidFunction, onRemove: VoidFunction)}
	<div class="border p-2">
		<Header title={label} subheading={description}>
			{#snippet actions()}
				<Button size="sm" iconOnly icon={mdiPencil} onclick={onEdit} />
				<Button size="sm" iconOnly icon={mdiMinus} onclick={onRemove} />
			{/snippet}
		</Header>
	</div>
{/snippet}
