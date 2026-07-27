<script lang="ts">
	import { Handle, Position, type Node, type NodeProps } from "@xyflow/svelte";
	import { Badge } from "$components/ui/badge";
	import { useSystemMapViewController, type SystemMapNodeData } from "./controller.svelte";

	let { data }: NodeProps<Node<SystemMapNodeData>> = $props();
	const view = useSystemMapViewController();

	const state = $derived(data.entity.attributes.latestEvidence?.attributes.subjectState);
	const aliases = $derived(data.entity.attributes.aliases);
	const label = $derived(
		state?.displayName || aliases[0]?.attributes.providerSubjectRef || "Unknown entity"
	);
</script>

<Handle type="target" position={Position.Left} />
<button
	type="button"
	class="border-border bg-card text-card-foreground hover:border-primary/60 flex w-56 flex-col items-start gap-1 border px-3 py-2 text-left shadow-sm transition-colors"
	onclick={() => view.selectEntity(data.entity)}
>
	<Badge variant="outline" class="max-w-full truncate text-xs">
		{data.entity.attributes.kind.replaceAll("_", " ")}
	</Badge>
	<span class="w-full truncate text-sm font-medium">{label}</span>
	{#if state?.description}
		<span class="text-muted-foreground line-clamp-2 text-xs">
			{state.description}
		</span>
	{/if}
</button>
<Handle type="source" position={Position.Right} />
