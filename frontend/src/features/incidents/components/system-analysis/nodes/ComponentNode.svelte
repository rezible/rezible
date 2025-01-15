<script lang="ts">
	import {
		NodeToolbar,
		type NodeProps,
		Handle,
		Position,
	} from "@xyflow/svelte";
    import { Button, ButtonGroup } from "svelte-ux";
    import type { SystemComponentNodeData } from "../diagram.svelte";

	const { data: arbitraryData }: NodeProps = $props();
	const data = $derived(arbitraryData as SystemComponentNodeData);
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

<div class="node border bg-surface-100 rounded-lg p-3 group">
	<span>{data.component.attributes.name}</span>
	<Handle type="target" position={Position.Left} class="invisible group-hover:visible" />
	<Handle type="source" position={Position.Right} class="invisible group-hover:visible" />
</div>
