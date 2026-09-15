<script lang="ts">
	import { onMount, type Snippet } from "svelte";
	import {
		Background,
		BackgroundVariant,
		ControlButton,
		Controls,
		MiniMap,
		SvelteFlow,
		SvelteFlowProvider,
		type NodeTypes,
		type EdgeTypes,
	} from "@xyflow/svelte";
	import RiFullscreenLine from "remixicon-svelte/icons/fullscreen-line";
	import "@xyflow/svelte/dist/style.css";
	import { mode } from "mode-watcher";
	import type { SystemDiagramController } from "./controller.svelte";
	import EntityNode from "./entity-node/EntityNode.svelte";
	import RelationshipEdge from "./relationship-edge/RelationshipEdge.svelte";

	type Props = {
		controller: SystemDiagramController;
		nodeTypes?: NodeTypes;
		edgeTypes?: EdgeTypes;
		children?: Snippet;
	};
	let { controller, nodeTypes, edgeTypes, children }: Props = $props();

	onMount(() => {
		const mountedController = controller;
		return () => mountedController.detach();
	});
</script>

<SvelteFlowProvider>
	<div
		class="relative h-full min-h-0 w-full min-w-0"
		role="presentation"
		bind:clientWidth={controller.width}
		bind:clientHeight={controller.height}
		onkeydowncapture={controller.keydown}
	>
		<SvelteFlow
			class="rezible-flow"
			nodeTypes={{ entity: EntityNode, ...nodeTypes }}
			edgeTypes={{ relationship: RelationshipEdge, ...edgeTypes }}
			colorMode={mode.current === "dark" ? "dark" : "light"}
			proOptions={{ hideAttribution: true }}
			minZoom={0.01}
			bind:nodes={controller.nodes}
			bind:edges={controller.edges}
			bind:viewport={controller.viewport}
			nodesConnectable={false}
			elementsSelectable={false}
			disableKeyboardA11y
			deleteKey={null}
			multiSelectionKey={null}
			selectionKey={null}
			oninit={controller.attach}
			onnodeclick={({ node, event }) => controller.select({ nodeId: node.id }, event)}
			onedgeclick={({ edge, event }) => controller.select({ edgeId: edge.id }, event)}
			onpaneclick={(event) => controller.select({}, event.event)}
			onnodedragstop={({ targetNode }) => {
				if (targetNode) controller.moveNode(targetNode.id, targetNode.position);
			}}
			onnodecontextmenu={({ node, event }) => controller.contextMenu({ nodeId: node.id }, event)}
			onedgecontextmenu={({ edge, event }) => controller.contextMenu({ edgeId: edge.id }, event)}
		>
			<Background variant={BackgroundVariant.Dots} />
			<Controls position="top-left" showFitView={false} showLock={false}>
				<ControlButton title="Fit diagram" aria-label="Fit diagram" onclick={controller.fit}>
					<RiFullscreenLine />
				</ControlButton>
			</Controls>
			<MiniMap position="top-right" />
			{@render children?.()}
		</SvelteFlow>
	</div>
</SvelteFlowProvider>
