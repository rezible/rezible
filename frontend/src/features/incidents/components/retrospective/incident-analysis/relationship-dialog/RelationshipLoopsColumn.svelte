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
	import { Header } from "svelte-ux";

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
</script>

<div class="border">
	<Header title="Source" />

	<Header title="Feedback Signals" />
	{#each sourceFeedbackSignals as {feedback, signal}}
		<div class="border p-2">
			{signal.attributes.label} - {feedback.attributes.description}
		</div>
	{/each}

	<Header title="Control Actions" />
	{#each sourceControlActions as {action, control}}
		<div class="border p-2">
			{control.attributes.label} - {action.attributes.description}
		</div>
	{/each}
</div>

<div class="border">
	<Header title="Target" />

	<Header title="Feedback Signals" />
	{#each targetFeedbackSignals as {feedback, signal}}
		<div class="border p-2">
			{signal.attributes.label} - {feedback.attributes.description}
		</div>
	{/each}

	<Header title="Control Actions" />
	{#each targetControlActions as ca}
		{@render controlActionItem(ca)}
	{/each}
</div>

{#snippet controlActionItem(ca: RelationshipControlAction)}
<div class="border p-2">
	{ca.control.attributes.label} - {ca.action.attributes.description}
</div>
{/snippet}