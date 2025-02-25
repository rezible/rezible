<script lang="ts">
	import { createQuery } from "@tanstack/svelte-query";
	import { getSystemComponentOptions, type SystemComponent } from "$lib/api";
	import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";
	import { relationshipAttributes } from "./dialogState.svelte";
	import RelationshipAttributesComponentColumn from "./RelationshipAttributesComponentColumn.svelte";
	import RelationshipLoopsColumn from "./RelationshipLoopsColumn.svelte";

	const sourceId = $derived(relationshipAttributes.sourceId);
	const sourceComponentQuery = createQuery(() => ({
		...getSystemComponentOptions({path: {id: sourceId}}),
		enabled: !!sourceId,
	}));
	const sourceComponent = $derived(sourceComponentQuery.data?.data);

	const targetId = $derived(relationshipAttributes.targetId);
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