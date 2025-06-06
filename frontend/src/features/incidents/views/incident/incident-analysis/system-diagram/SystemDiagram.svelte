<script lang="ts">
	import {
		SvelteFlowProvider,
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

	import { settings } from "$lib/settings.svelte";

	import { SystemDiagramState, setSystemDiagram } from "./diagramState.svelte";

	import ContextMenu from "./ContextMenu.svelte";
	import ConnectionLine from "./ConnectionLine.svelte";
	import ActionsBar from "./SystemDiagramActionsBar.svelte";
	import EditToolbar from "./EditToolbar.svelte";
	import AddingComponentGhostNode from "./AddingComponentGhostNode.svelte";
	import ComponentNode from "./ComponentNode.svelte";
	import RelationshipEdge from "./RelationshipEdge.svelte";
	import ComponentDialog from "./component-dialog/ComponentDialog.svelte";
	import RelationshipDialog from "./relationship-dialog/RelationshipDialog.svelte";
	import { ComponentDialogState, setComponentDialog } from "./component-dialog/dialogState.svelte";
	import { RelationshipDialogState, setRelationshipDialog } from "./relationship-dialog/dialogState.svelte";

	setComponentDialog(new ComponentDialogState());
	setRelationshipDialog(new RelationshipDialogState());

	let containerEl = $state<HTMLElement>();
	const diagram = new SystemDiagramState(() => containerEl);
	setSystemDiagram(diagram);

	const colorMode = $derived<ColorMode>(settings.theme.dark ? "dark" : "light");

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

<SvelteFlowProvider>
	<div
		class="h-full w-full overflow-hidden relative"
		role="presentation"
		bind:this={containerEl}
		oncontextmenu={(e) => e.preventDefault()}
	>
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
			{#if diagram.ctxMenuProps}
				<ContextMenu />
			{/if}
			<EditToolbar />
			<AddingComponentGhostNode />

			<Panel position="bottom-right">
				<ActionsBar />
			</Panel>
		</SvelteFlow>
	</div>

	<ComponentDialog />
	<RelationshipDialog />
</SvelteFlowProvider>