<script lang="ts">
	import type { Edge, Node } from "@xyflow/svelte";
	import { useSystemAnalysisController } from "../../controller.svelte";

	type Props = { selected: { node?: Node; edge?: Edge } };
	let { selected }: Props = $props();
	
	const analysis = useSystemAnalysisController();
	const entries = $derived(
		selected.node
			? (analysis.attachments.byNodeId.get(selected.node.id) ?? [])
			: selected.edge
				? (analysis.attachments.byEdgeId.get(selected.edge.id) ?? [])
				: []
	);
</script>

{#if selected.node || selected.edge}
	<aside class="w-72 max-h-80 overflow-auto border border-border bg-card p-3 text-card-foreground shadow">
		<h3 class="mb-2 text-sm font-semibold">{selected.node ? "Node entries" : "Relationship entries"}</h3>
		{#if entries.length === 0}
			<p class="text-xs text-muted-foreground">No entries attached.</p>
		{:else}
			<div class="space-y-2">
				{#each entries as entry (entry.id)}
					<article class="border-l-2 border-primary pl-2 text-xs">
						<div class="text-muted-foreground">
							{entry.attributes.kind}{entry.attributes.occurredAt
								? ` · ${new Date(entry.attributes.occurredAt).toLocaleString()}`
								: ""}
						</div>
						<div class="font-medium">{entry.attributes.title}</div>
						{#if entry.attributes.body}<p class="mt-1 whitespace-pre-wrap">
								{entry.attributes.body}
							</p>{/if}
					</article>
				{/each}
			</div>
		{/if}
	</aside>
{/if}
