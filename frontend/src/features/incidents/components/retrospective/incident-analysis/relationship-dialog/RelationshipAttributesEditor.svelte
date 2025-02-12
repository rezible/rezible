<script lang="ts">
	import {
	type SystemAnalysisRelationshipControlAction,
	type SystemAnalysisRelationshipFeedbackSignal,
		type SystemComponent,
		type SystemComponentControl,
		type SystemComponentSignal,
		getSystemComponentOptions,
	} from "$lib/api";
	import { relationshipDialog } from "./relationshipDialog.svelte";
	import { createQuery } from "@tanstack/svelte-query";
	import LoadingQueryWrapper from "$src/components/loader/LoadingQueryWrapper.svelte";
	import RelationshipAttributesComponentColumn from "./RelationshipAttributesComponentColumn.svelte";
	import { Button, Header } from "svelte-ux";
	import { SvelteMap, SvelteSet } from "svelte/reactivity";
	import SystemComponents from "../incident-timeline/event-editor/panels/SystemComponents.svelte";
	import RelationshipLoopsColumn from "./RelationshipLoopsColumn.svelte";

	const attrs = $derived(relationshipDialog.attributes);

	const sourceId = $derived(attrs.sourceId);
	const sourceComponentQuery = createQuery(() => ({
		...getSystemComponentOptions({path: {id: sourceId}}),
		enabled: !!sourceId,
	}));
	const sourceComponent = $derived(sourceComponentQuery.data?.data);

	const targetId = $derived(attrs.targetId);
	const targetComponentQuery = createQuery(() => ({
		...getSystemComponentOptions({path: {id: targetId}}),
		enabled: !!targetId,
	}));
	const targetComponent = $derived(targetComponentQuery.data?.data);
</script>

<div class="grid grid-cols-3 min-h-0 max-h-full h-full gap-2 overflow-y-hidden">
	<div class="border overflow-y-auto">
		<LoadingQueryWrapper query={sourceComponentQuery}>
			{#snippet view(component: SystemComponent)}
				<RelationshipAttributesComponentColumn {component} />
			{/snippet}
		</LoadingQueryWrapper>
	</div>

	<div class="border p-2 overflow-y-auto">
		{#if !!sourceComponent && !!targetComponent}
			<RelationshipLoopsColumn source={sourceComponent} target={targetComponent} />
		{/if}
	</div>

	<div class="border overflow-y-auto">
		<LoadingQueryWrapper query={targetComponentQuery}>
			{#snippet view(component: SystemComponent)}
				<RelationshipAttributesComponentColumn {component} />
			{/snippet}
		</LoadingQueryWrapper>
	</div>
</div>