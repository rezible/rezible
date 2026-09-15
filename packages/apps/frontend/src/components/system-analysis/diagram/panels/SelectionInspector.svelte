<script lang="ts">
	import { useDiagramController } from "../diagramController.svelte";
	import { useSystemAnalysisController } from "../../controller.svelte";

	const diagram = useDiagramController();
	const node = $derived(diagram.selectedNode);
	const edge = $derived(diagram.selectedEdge);

	const analysis = useSystemAnalysisController();
	const nodeEntries = $derived(!!node ? (analysis.attachments.byNodeId.get(node.id) ?? []) : []);
	const edgeEntries = $derived(!!edge ? (analysis.attachments.byEdgeId.get(edge.id) ?? []) : []);
	const entries = $derived(nodeEntries || edgeEntries || []);
</script>

{#if node || edge}
	<aside class="w-72 max-h-80 overflow-auto border border-border bg-card p-3 text-card-foreground shadow">
		<h3 class="mb-2 text-sm font-semibold">{node ? "Node entries" : "Relationship entries"}</h3>
		{#if entries.length === 0}
			<p class="text-xs text-muted-foreground">No entries attached.</p>
		{:else}
			<div class="space-y-2">
				{#each entries as entry (entry.id)}
					{@const attrs = entry.attributes}
					<article class="border-l-2 border-primary pl-2 text-xs">
						<div class="text-muted-foreground">
							{attrs.kind}
							{attrs.occurredAt ? ` · ${new Date(attrs.occurredAt).toLocaleString()}` : ""}
						</div>
						<div class="font-medium">{attrs.title}</div>
						{#if attrs.body}
							<p class="mt-1 whitespace-pre-wrap">
								{attrs.body}
							</p>
						{/if}
					</article>
				{/each}
			</div>
		{/if}
	</aside>
{/if}
