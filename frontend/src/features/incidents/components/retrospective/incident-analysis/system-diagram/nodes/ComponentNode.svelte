<script lang="ts">
	import {
		NodeToolbar,
		type NodeProps,
		Handle,
		Position,
        useNodes,
	} from "@xyflow/svelte";
    import { Button, ButtonGroup } from "svelte-ux";
    import type { SystemComponentNodeData } from "../diagram.svelte";

	const { 
		id,
		selected,
		data: arbitraryData,
	}: NodeProps = $props();
	const data = $derived(arbitraryData as SystemComponentNodeData);

	const nodes = useNodes();
</script>

<NodeToolbar position={Position.Bottom}>
	<div class="border rounded-lg bg-surface-100 p-2">
		<ButtonGroup variant="fill-light" color="accent">
			<Button>delete</Button>
			<Button>copy</Button>
			<Button>expand</Button>
		</ButtonGroup>
	</div>
</NodeToolbar>

<div data-is-selected={selected} class="node border bg-surface-100 data-[is-selected=true]:bg-surface-200 rounded-lg p-3 group">
	<span>{data.component.attributes.name}</span>
	<Handle type="target" position={Position.Left} class="invisible group-hover:visible" />
	<Handle type="source" position={Position.Right} class="invisible group-hover:visible" />
</div>
