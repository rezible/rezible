<script lang="ts">
	import { BaseEdge, getBezierPath, type EdgeProps } from "@xyflow/svelte";
	import { cn } from "$lib/utils";
	import type { SystemDiagramEdge } from "../types";

	type Props = EdgeProps<SystemDiagramEdge>;
	let props: Props = $props();

	const path = $derived(getBezierPath(props));
	const data = $derived(props.data);

	const relationship = $derived(data?.relationship);
	const attrs = $derived(relationship?.attributes);
	const label = $derived(attrs?.latestState?.displayName || attrs?.predicate.replaceAll("_", " "));
</script>

<BaseEdge
	id={props.id}
	path={path[0]}
	markerEnd={props.markerEnd}
	interactionWidth={24}
	class={cn(data?.highlighted && "stroke-primary")}
/>

<text
	x={path[1]}
	y={path[2]}
	text-anchor="middle"
	dominant-baseline="central"
	stroke-width="4"
	class="fill-foreground stroke-background text-xs [paint-order:stroke]"
>
	{label}
</text>
