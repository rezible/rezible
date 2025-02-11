<script lang="ts" module>
	type RelationshipFeedbackSignal = {
		feedback: SystemAnalysisRelationshipFeedbackSignal;
		signal: SystemComponentSignal;
	}

	type RelationshipControlAction = {
		action: SystemAnalysisRelationshipControlAction;
		control: SystemComponentControl;
	}

</script>

<script lang="ts">
	import type { SystemAnalysisRelationshipControlAction, SystemAnalysisRelationshipFeedbackSignal, SystemComponent, SystemComponentControl, SystemComponentSignal } from "$lib/api";
	import { SvelteMap } from "svelte/reactivity";
	import { relationshipDialog } from "./relationshipDialog.svelte";
	import { Button, Header } from "svelte-ux";
	import LabelDescriptionEditor from "./LabelDescriptionEditor.svelte";

	type Props = {
		source: SystemComponent;
		target: SystemComponent;
	}
	const { source, target }: Props = $props();

	const attrs = $derived(relationshipDialog.attributes);

	const sourceSignals = $derived(new SvelteMap(source.attributes.signals.map(s => [s.id, s])));
	const targetSignals = $derived(new SvelteMap(target.attributes.signals.map(s => [s.id, s])));

	const sourceControls = $derived(new SvelteMap(source.attributes.controls.map(s => [s.id, s])));
	const targetControls = $derived(new SvelteMap(target.attributes.controls.map(s => [s.id, s])));

	const [sourceFeedbackSignals, targetFeedbackSignals] = $derived.by(() => {
		const src: RelationshipFeedbackSignal[] = [];
		const tgt: RelationshipFeedbackSignal[] = [];
		attrs.feedbackSignals.forEach(feedback => {
			const signalId = feedback.attributes.signal_id;
			const sourceSignal = sourceSignals.get(signalId);
			if (sourceSignal) src.push({feedback, signal: sourceSignal});
			const targetSignal = targetSignals.get(signalId);
			if (targetSignal) tgt.push({feedback, signal: targetSignal});
		})
		return [src, tgt];
	});

	const [sourceControlActions, targetControlActions] = $derived.by(() => {
		const src: RelationshipControlAction[] = [];
		const tgt: RelationshipControlAction[] = [];
		attrs.controlActions.forEach(action => {
			const controlId = action.attributes.control_id;
			const sourceControl = sourceControls.get(controlId);
			if (sourceControl) src.push({action, control: sourceControl});
			const targetControl = targetControls.get(controlId);
			if (targetControl) tgt.push({action, control: targetControl});
		});
		return [src, tgt];
	});

	let editingAction = $state<SystemAnalysisRelationshipControlAction>();
	const updateEditingAction = () => {
		if (!editingAction) return;
		attrs.setControlAction($state.snapshot(editingAction.attributes));
		editingAction = undefined;
	} 

	let editingFeedback = $state<SystemAnalysisRelationshipFeedbackSignal>();
	const updateEditingFeedback = () => {
		if (!editingFeedback) return;
		attrs.setFeedbackSignal($state.snapshot(editingFeedback.attributes));
		editingFeedback = undefined;
	} 
</script>

<div class="flex flex-col gap-2">
	{@render componentFeedbackLoops(source.attributes.name, sourceFeedbackSignals, sourceControlActions)}
	{@render componentFeedbackLoops(target.attributes.name, targetFeedbackSignals, targetControlActions)}
</div>

{#snippet componentFeedbackLoops(title: string, feedbacks: RelationshipFeedbackSignal[], actions: RelationshipControlAction[])}
	<div class="border p-1">
		<Header {title} />

		{#if feedbacks.length === 0}
			<Header title="Supplies No Feedback" />
		{:else}
			{@render feedbackSignalItems(feedbacks)}
		{/if}

		{#if actions.length === 0}
			<Header title="Exposes No Controls" />
		{:else}
			{@render controlActionItems(actions)}
		{/if}
	</div>
{/snippet}

{#snippet feedbackSignalItems(feedbacks: RelationshipFeedbackSignal[])}
	<Header title="Sends Feedback" />
	{#if editingFeedback}
		{@const signalId = editingFeedback.attributes.signal_id}
		{@const ctrl = sourceSignals.get(signalId) || targetSignals.get(signalId)}
		<span>{ctrl?.attributes.label ?? "Editing"}</span>
		<LabelDescriptionEditor 
			bind:description={editingFeedback.attributes.description}
			onCancel={() => {editingFeedback = undefined}}
			onConfirm={updateEditingFeedback}
		/>
	{:else}
		{#each feedbacks as {feedback, signal}}
		<div class="border p-2">
			{signal.attributes.label} - {feedback.attributes.description}
			<Button on:click={() => {editingFeedback = $state.snapshot(feedback)}}>edit</Button>
			<Button on:click={() => {attrs.removeFeedbackSignal(feedback.id)}}>x</Button>
		</div>
		{/each}
	{/if}
{/snippet}

{#snippet controlActionItems(actions: RelationshipControlAction[])}
	<Header title="Is Controlled With" />

	{#if editingAction}
		{@const controlId = editingAction.attributes.control_id}
		{@const ctrl = sourceControls.get(controlId) || targetControls.get(controlId)}
		<span>{ctrl?.attributes.label ?? "Editing"}</span>
		<LabelDescriptionEditor 
			bind:description={editingAction.attributes.description}
			onCancel={() => {editingAction = undefined}}
			onConfirm={updateEditingAction}
		/>
	{:else}
		{#each actions as {control, action}}
			<div class="border p-2">
				{control.attributes.label} - {action.attributes.description}
				<Button on:click={() => {editingAction = $state.snapshot(action)}}>edit</Button>
				<Button on:click={() => {attrs.removeControlAction(action.id)}}>x</Button>
			</div>
		{/each}
	{/if}
{/snippet}