<script lang="ts">
	import { onMount } from "svelte";
	import { useSvelteFlow } from "@xyflow/svelte";

	import type { FlowEdge, FlowNode } from "../flow-model";
	import type { SystemMapController } from "../controller.svelte";

	type Props = { controller: SystemMapController };
	let { controller }: Props = $props();

	const flow = useSvelteFlow<FlowNode, FlowEdge>();
	onMount(() => {
		controller.setViewportApi((vp) => flow.setViewport(vp, { duration: 0 }));
		return () => {
			controller.setViewportApi(undefined);
		}
	});
</script>
