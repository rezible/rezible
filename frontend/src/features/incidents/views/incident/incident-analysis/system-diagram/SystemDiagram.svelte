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
		nodeTypes: {
			// @ts-expect-error this will be resolved
			default: ComponentNode,
			// @ts-expect-error this will be resolved
			component: ComponentNode,
		},
		edgeTypes: {
			// @ts-expect-error this will be resolved
			default: RelationshipEdge,
			// @ts-expect-error this will be resolved
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
			nodes={diagram.nodes}
			edges={diagram.edges}
			oninit={() => diagram.onFlowInit()}
			onconnect={e => diagram.onEdgeConnect(e)}
			on:panecontextmenu={e => diagram.handleContextMenuEvent(e)}
			on:edgecontextmenu={e => diagram.handleContextMenuEvent(e)}
			on:nodecontextmenu={e => diagram.handleContextMenuEvent(e)}
			on:selectioncontextmenu={e => diagram.handleContextMenuEvent(e)}
			on:nodeclick={e => diagram.handleNodeClicked(e)}
			on:nodedragstart={e => diagram.handleNodeDragStart(e)}
			on:nodedrag={e => diagram.handleNodeDrag(e)}
			on:nodedragstop={e => diagram.handleNodeDragStop(e)}
			on:paneclick={e => diagram.handlePaneClicked(e)}
			on:edgeclick={e => diagram.handleEdgeClicked(e)}
		>
			<Background {...backgroundSettings} />
			<Controls {...controlsSettings} />
			<MiniMap {...minimapSettings} />
			<ConnectionLine slot="connectionLine" />
			<ContextMenu />
			<EditToolbar />
			<AddingComponentGhostNode />
		</SvelteFlow>

		<ActionsBar />
	</div>

	<ComponentDialog />
	<RelationshipDialog />
</SvelteFlowProvider>