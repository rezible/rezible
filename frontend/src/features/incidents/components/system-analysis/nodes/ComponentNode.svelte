<script lang="ts">
	import {
		NodeToolbar,
		type NodeProps,
		Handle,
		Position,
	} from "@xyflow/svelte";
    import { Button, ButtonGroup } from "svelte-ux";

	const { data }: NodeProps = $props();
</script>

<NodeToolbar>
	<div class="border rounded-lg bg-surface-100/20 p-2">
		<ButtonGroup variant="fill-light" color="neutral">
			<Button>delete</Button>
			<Button>copy</Button>
			<Button>expand</Button>
		</ButtonGroup>
	</div>
</NodeToolbar>

<div class="node border rounded-lg p-3">
	<span>component: {data.label}</span>
	<Handle type="target" position={Position.Left} />
	<Handle type="source" position={Position.Right} />
</div>
