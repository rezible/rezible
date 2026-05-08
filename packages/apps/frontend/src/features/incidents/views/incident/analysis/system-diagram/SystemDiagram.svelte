<script lang="ts">
	import {
		SvelteFlow,
		Controls,
		Background,
		BackgroundVariant,
		MiniMap,
		type BackgroundProps,
		type ControlsProps,
		type MiniMapProps,
		type SvelteFlowProps,
		type ColorMode,
		type Viewport,
		Panel,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";
	import "./diagram-styles.css";

	import { useSystemDiagram } from "./diagramState.svelte";

	import SystemDiagramContextMenu from "./SystemDiagramContextMenu.svelte";
	import ConnectionLine from "./ConnectionLine.svelte";
	import ActionsBar from "./SystemDiagramActionsBar.svelte";
	import EditToolbar from "./EditToolbar.svelte";
	import AddingEntityGhostNode from "./AddingEntityGhostNode.svelte";
	import ComponentNode from "./ComponentNode.svelte";
	import RelationshipEdge from "./RelationshipEdge.svelte";
	import { useIncidentAnalysis } from "../controller.svelte";

	const analysis = useIncidentAnalysis();
	const diagram = useSystemDiagram();

	const colorMode = $derived<ColorMode>("dark");

	const flowSettings: SvelteFlowProps = {
		connectionLineComponent: ConnectionLine,
		nodeTypes: {
			default: ComponentNode,
			component: ComponentNode,
		},
		edgeTypes: {
			default: RelationshipEdge,
			relationship: RelationshipEdge,
		},
		snapGrid: [25, 25],
		connectionRadius: 40,
		fitView: true,
		proOptions: { hideAttribution: true },
	};

	const backgroundSettings: BackgroundProps = {
		variant: BackgroundVariant.Dots,
	};

	const controlsSettings: ControlsProps = {
		position: "top-left",
	};

	const minimapSettings: MiniMapProps = {
		position: "top-right",
	};

	let viewport = $state<Viewport>({ x: 100, y: 100, zoom: 1.25 });
</script>

<SvelteFlow
	{...flowSettings}
	{colorMode}
	bind:nodes={diagram.nodes}
	bind:edges={diagram.edges}
	bind:viewport
	oninit={() => diagram.onFlowInit()}
	onconnect={e => diagram.onEdgeConnect(e)}
	onpanecontextmenu={e => diagram.handleContextMenuEvent(e)}
	onedgecontextmenu={e => diagram.handleContextMenuEvent(e)}
	onnodecontextmenu={e => diagram.handleContextMenuEvent(e)}
	onselectioncontextmenu={e => diagram.handleContextMenuEvent(e)}
	onnodeclick={e => diagram.handleNodeClicked(e)}
	onnodedragstart={e => diagram.handleNodeDragStart(e)}
	onnodedrag={e => diagram.handleNodeDrag(e)}
	onnodedragstop={e => diagram.handleNodeDragStop(e)}
	onpaneclick={e => diagram.handlePaneClicked(e)}
	onedgeclick={e => diagram.handleEdgeClicked(e)}
>
	<Background {...backgroundSettings} />
	<Controls {...controlsSettings} />
	<MiniMap {...minimapSettings} />
	{#if !!analysis.contextMenu.diagram}
		<SystemDiagramContextMenu {...analysis.contextMenu.diagram} />
	{/if}
	<EditToolbar />
	<AddingEntityGhostNode />

	<Panel position="bottom-right">
		<ActionsBar />
	</Panel>
</SvelteFlow>
