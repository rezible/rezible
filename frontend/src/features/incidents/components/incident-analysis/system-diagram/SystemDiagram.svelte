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
		type NodeTypes,
		type EdgeTypes,
		type ColorMode,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

	import { settings } from "$lib/settings.svelte";

	import { diagram } from "./diagram.svelte";
	import ContextMenu from "./ContextMenu.svelte";
	import ConnectionLine from "./ConnectionLine.svelte";
	import ActionsBar from "./SystemDiagramActionsBar.svelte";
	import EditToolbar from "./EditToolbar.svelte";
	import AddingComponentGhostNode from "./AddingComponentGhostNode.svelte";
	import ComponentNode from "./ComponentNode.svelte";
	import RelationshipEdge from "./RelationshipEdge.svelte";
	import ComponentDialog from "./component-dialog/ComponentDialog.svelte";
	import RelationshipDialog from "./relationship-dialog/RelationshipDialog.svelte";

	type Props = {};
	const {}: Props = $props();

	let containerEl = $state<HTMLElement>();
	diagram.setup(() => containerEl);

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
			oninit={diagram.onFlowInit}
			onconnect={diagram.onEdgeConnect}
			on:panecontextmenu={diagram.handleContextMenuEvent}
			on:edgecontextmenu={diagram.handleContextMenuEvent}
			on:nodecontextmenu={diagram.handleContextMenuEvent}
			on:selectioncontextmenu={diagram.handleContextMenuEvent}
			on:nodeclick={diagram.handleNodeClicked}
			on:nodedragstart={diagram.handleNodeDragStart}
			on:nodedrag={diagram.handleNodeDrag}
			on:paneclick={diagram.handlePaneClicked}
			on:edgeclick={diagram.handleEdgeClicked}
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
</SvelteFlowProvider>

<ComponentDialog />

<RelationshipDialog />
