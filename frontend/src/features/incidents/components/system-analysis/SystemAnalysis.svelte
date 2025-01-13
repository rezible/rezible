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
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

    import { diagram } from "./diagram.svelte";
	import ContextMenu from "./SystemDiagramContextMenu.svelte";
    import AnalysisToolbar from "./AnalysisToolbar.svelte";
    import ComponentNode from "./nodes/ComponentNode.svelte";
    import type { SystemComponentAttributes } from "$src/lib/api";
    import type { Component } from "svelte";

	type Props = {}
	const {  }: Props = $props();

	let containerEl = $state<HTMLElement>();
	diagram.componentSetup(() => containerEl);

	const nodeTypes: Record<SystemComponentAttributes["kind"], Component<NodeProps>> = {
		service: ComponentNode,
	}

	const flowProps = $derived<SvelteFlowProps>({
		// @ts-expect-error this will be resolved
		nodeTypes,
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

	}

	const minimapProps: MiniMapProps = {

	}
</script>

<div class="flex flex-col gap-2 h-full w-full overflow-y-hidden">
	<SvelteFlowProvider>
		<div class="h-fit w-full">
			<AnalysisToolbar />
		</div>
		<div class="min-h-0 flex-1" role="presentation" bind:this={containerEl} oncontextmenu={e => e.preventDefault()}>
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
		</div>
	</SvelteFlowProvider>
</div>
