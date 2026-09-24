<script lang="ts">
	import { onMount } from "svelte";
	import { useSvelteFlow, type FitViewOptions } from "@xyflow/svelte";
	import type { FlowNode, FlowEdge } from "./flow-graph-model";
	import type { SystemMapController } from "./controller.svelte";

	type Props = { controller: SystemMapController };
	let { controller }: Props = $props();
	const flow = useSvelteFlow<FlowNode, FlowEdge>();
	onMount(() => {
		controller.setViewportApi({
			fit: async (ids, request) => {
				const options: FitViewOptions<FlowNode> = {
					padding: 0.2,
					minZoom: 0.1,
					maxZoom: 1,
					duration: 0,
				};

				if (ids) {
					options.nodes = ids.map((id) => ({ id }));
				}

				if (request?.preserveZoom) {
					const zoom = flow.getViewport().zoom;
					options.minZoom = zoom;
					options.maxZoom = zoom;
				}

				await flow.fitView(options);
			},
		});

		return () => {
			controller.setViewportApi(undefined);
		};
	});
</script>
