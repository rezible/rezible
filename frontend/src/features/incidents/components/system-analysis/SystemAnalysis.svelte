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
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

	import { diagram } from "./diagram.svelte";
	import ContextMenu from "./SystemDiagramContextMenu.svelte";
	import SystemDiagramToolbar from "./SystemDiagramToolbar.svelte";
	import ComponentNode from "./nodes/ComponentNode.svelte";
	import RelationshipEdge from "./edges/RelationshipEdge.svelte";
    import SystemDiagramConnectionLine from "./SystemDiagramConnectionLine.svelte";

	type Props = {}
	const {}: Props = $props();

	let containerEl = $state<HTMLElement>();
	diagram.setup(() => containerEl);

	const colorMode = $derived("dark"); // get from svelte-ux theme

	const nodeTypes: NodeTypes = {
		// @ts-expect-error this will be resolved
		default: ComponentNode, component: ComponentNode,
	}

	const edgeTypes: EdgeTypes = {
		// @ts-expect-error this will be resolved
		default: RelationshipEdge, relationship: RelationshipEdge,
	}

	const flowProps = $derived<SvelteFlowProps>({
		nodeTypes, 
		edgeTypes,
		nodes: diagram.nodes,
		edges: diagram.edges,
		colorMode,
		snapGrid: [25, 25],
		fitView: true,
		proOptions: {hideAttribution: true}
	});

	const backgroundProps: BackgroundProps = {
		variant: BackgroundVariant.Dots,
	};

	const controlsProps: ControlsProps = {
		position: "top-left",
	}

	const minimapProps: MiniMapProps = {
		position: "top-right",
	}
</script>

<div class="h-full w-full overflow-hidden relative" role="presentation" bind:this={containerEl} oncontextmenu={e => e.preventDefault()}>
	<SvelteFlowProvider>
		<SvelteFlow
			{...flowProps}
			on:panecontextmenu={diagram.handleContextMenuEvent}
			on:edgecontextmenu={diagram.handleContextMenuEvent}
			on:nodecontextmenu={diagram.handleContextMenuEvent}
			on:selectioncontextmenu={diagram.handleContextMenuEvent}
			on:nodeclick={diagram.handleNodeClicked}
			on:paneclick={diagram.handlePaneClicked}
			on:edgeclick={e => console.log("edge click", e)}
		>
			<Background {...backgroundProps} />
			<Controls {...controlsProps} />
			<MiniMap {...minimapProps} />
			<SystemDiagramConnectionLine slot="connectionLine" />
			<ContextMenu {...diagram.ctxMenuProps} />
		</SvelteFlow>
		<SystemDiagramToolbar />
	</SvelteFlowProvider>
</div>
