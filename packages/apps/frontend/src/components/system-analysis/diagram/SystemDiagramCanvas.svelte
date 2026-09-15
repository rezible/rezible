<script lang="ts">
	import {
		SvelteFlow,
		Controls,
		Background,
		BackgroundVariant,
		MiniMap,
		type Viewport,
		Panel,
		useSvelteFlow,
		useStore,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

	import { useDiagramController } from "./diagramController.svelte";

	import {
		ComponentNode,
		RelationshipEdge,
		ConnectionLine,
		AddingEntityGhostNode,
	} from "./topology-components";

	import { ContextMenu, EditToolbar, ActionsBar } from "./panels";
	import { mode } from "mode-watcher";

	const diagram = useDiagramController();
	const flow = useSvelteFlow();
	const flowStore = useStore();

	let viewport = $state<Viewport>({ x: 100, y: 100, zoom: 1.25 });
</script>

<SvelteFlow
	class="rezible-flow"
	connectionLineComponent={ConnectionLine}
	nodeTypes={{default: ComponentNode, component: ComponentNode}}
	edgeTypes={{default: RelationshipEdge, relationship: RelationshipEdge}}
	snapGrid={[25, 25]}
	connectionRadius={40}
	fitView={true}
	proOptions={{ hideAttribution: true }}
	colorMode={mode.current === "dark" ? "dark" : "light"}
	bind:nodes={diagram.nodes}
	bind:edges={diagram.edges}
	bind:viewport
	oninit={() => diagram.onFlowInit(flow, flowStore)}
	onconnect={(e) => diagram.onEdgeConnect(e)}
	onpanecontextmenu={(e) => diagram.handleContextMenuEvent(e)}
	onedgecontextmenu={(e) => diagram.handleContextMenuEvent(e)}
	onnodecontextmenu={(e) => diagram.handleContextMenuEvent(e)}
	onselectioncontextmenu={(e) => diagram.handleContextMenuEvent(e)}
	onnodeclick={(e) => diagram.handleNodeClicked(e)}
	onnodedragstart={(e) => diagram.handleNodeDragStart(e)}
	onnodedrag={(e) => diagram.handleNodeDrag(e)}
	onnodedragstop={(e) => diagram.handleNodeDragStop(e)}
	onpaneclick={(e) => diagram.handlePaneClicked(e)}
	onedgeclick={(e) => diagram.handleEdgeClicked(e)}
>
	<Background variant={BackgroundVariant.Dots} />

	<Controls position={"top-left"} />

	<MiniMap position={"top-right"} />

	<ContextMenu />

	<EditToolbar />

	<AddingEntityGhostNode />

	<Panel position="bottom-right">
		<ActionsBar />
	</Panel>
</SvelteFlow>
