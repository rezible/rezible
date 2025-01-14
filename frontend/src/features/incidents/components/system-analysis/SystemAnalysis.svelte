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
        type NodeProps,
        type EdgeProps,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

    import { diagram } from "./diagram.svelte";
	import ContextMenu from "./SystemDiagramContextMenu.svelte";
    import AnalysisToolbar from "./AnalysisToolbar.svelte";
    import ComponentNode from "./nodes/ComponentNode.svelte";
    import type { SystemComponentAttributes, SystemComponentRelationship } from "$src/lib/api";
    import type { Component } from "svelte";
    import ControlEdge from "./edges/ControlEdge.svelte";
    import FeedbackEdge from "./edges/FeedbackEdge.svelte";

	type Props = {}
	const {  }: Props = $props();

	let containerEl = $state<HTMLElement>();
	diagram.setup(() => containerEl);

	const nodeTypes: Record<SystemComponentAttributes["kind"], Component<NodeProps>> = {
		service: ComponentNode,
	}

	const edgeTypes: Record<SystemComponentRelationship["attributes"]["kind"], Component<EdgeProps>> = {
		control: ControlEdge,
		feedback: FeedbackEdge,
	}

	const flowProps = $derived<SvelteFlowProps>({
		// @ts-expect-error this will be resolved
		nodeTypes, edgeTypes,
		nodes: diagram.nodes,
		edges: diagram.edges,
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
			<ContextMenu {...diagram.ctxMenuProps} />
		</SvelteFlow>
		<AnalysisToolbar />
	</SvelteFlowProvider>
</div>
