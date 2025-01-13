<script lang="ts">
	import {
		NodeToolbar,
		type NodeProps,
		Handle,
		Position,
	} from "@xyflow/svelte";

	const { data }: NodeProps = $props();
</script>

<NodeToolbar>
	<button>delete</button>
	<button>copy</button>
	<button>expand</button>
</NodeToolbar>

<div class="node">
	<div>component: {data.label}</div>
	<Handle type="target" position={Position.Left} />
	<Handle type="source" position={Position.Right} />
</div>
