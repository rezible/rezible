<script lang="ts">
	import {
		Background,
		Controls,
		Panel,
		SvelteFlow,
		SvelteFlowProvider,
		type NodeTypes,
		type EdgeTypes,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

	import { mode } from "mode-watcher";
	import * as Alert from "$components/ui/alert";
	import * as Empty from "$components/ui/empty";
	import { Button } from "$components/ui/button";

	import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";

	import MapConnection from "./map-connection/MapConnection.svelte";
	import MapNode from "./map-node/MapNode.svelte";
	import SystemMapViewport from "./SystemMapViewport.svelte";
	import { SystemMapController } from "./controller.svelte";
	import type { GraphSubset } from "$features/system/lib/system-map/graph";
	import type { Point } from "$features/system/lib/system-map/geometry";

	export type SystemMapProps = {
		graph: GraphSubset;
		/** Optional sparse world-space positions, applied exactly. */
		positions?: Readonly<Partial<Record<string, Point>>>;
		selection?: MapSelection;
		highlights?: MapHighlights;
		/** Emits the user's selection intention without owning an independent selection. */
		onSelectionChange: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
		/** Enables editing. Called once after a completed drag that changed the node position. */
		onNodeMove?: (entityId: string, position: Point) => void;
	};

	const { 
		graph, 
		positions, 
		selection, 
		highlights, 
		onSelectionChange, 
		onNodeMove,
	}: SystemMapProps = $props();

	const controller = new SystemMapController({
		graph: () => graph,
		positions: () => positions,
		selection: () => selection,
		highlights: () => highlights,
		onSelectionChange: (next, trigger) => onSelectionChange(next, trigger),
		onNodeMove: () => onNodeMove,
	});
	const nodeTypes: NodeTypes = { "system-map-node": MapNode };
	const edgeTypes: EdgeTypes = { "system-map-connection": MapConnection };

	export function fit() {
		controller.fit();
	}

	export function recenter() {
		controller.recenter();
	}

	export function reveal(target: MapSelection) {
		controller.reveal(target);
	}
</script>

<SvelteFlowProvider>
	<SystemMapViewport {controller} />
	<div class="system-map relative h-full min-h-0 w-full" onkeydowncapture={controller.keydown}>
		<div class="h-full w-full" class:invisible={controller.loading} aria-busy={controller.loading}>
			<SvelteFlow
				{nodeTypes}
				{edgeTypes}
				bind:nodes={controller.nodes}
				bind:edges={controller.edges}
				colorMode={mode.current === "dark" ? "dark" : "light"}
				proOptions={{ hideAttribution: true }}
				minZoom={0.1}
				maxZoom={2}
				nodesDraggable={controller.nodesDraggable}
				nodesConnectable={false}
				elementsSelectable={false}
				deleteKey={null}
				onnodeclick={({ node, event }) => controller.selectNode(node.id, event)}
				onedgeclick={({ edge, event }) => controller.selectEdge(edge.id, event)}
				onedgepointerenter={({ edge }) => controller.hoverEdge(edge.id)}
				onedgepointerleave={() => controller.hoverEdge(undefined)}
				onpaneclick={controller.clearSelection}
				onnodedragstart={controller.onNodeDragStart}
				onnodedragstop={controller.onNodeDragStop}
			>
				<Background />
				<Controls showFitView={false} showLock={false} />
				<Panel position="top-left">
					<div
						class="bg-background border-border flex items-center gap-1 rounded-md border p-1 shadow-sm"
					>
						<Button variant="outline" size="sm" onclick={controller.fit}>Fit</Button>
						<Button variant="outline" size="sm" onclick={controller.recenter}>Recenter</Button>
						<Button
							variant={controller.showLabels ? "secondary" : "outline"}
							size="sm"
							aria-pressed={controller.showLabels}
							onclick={controller.toggleLabels}
						>
							Labels
						</Button>
					</div>
				</Panel>
			</SvelteFlow>
		</div>
		{#if controller.loading}
			<div class="absolute inset-0 flex items-center justify-center" role="status">
				<span class="text-muted-foreground text-sm">Laying out map…</span>
			</div>
		{/if}
		{#if !controller.loading && !controller.nodes.length}
			<div class="pointer-events-none absolute inset-0 flex items-center justify-center">
				<Empty.Root>
					<Empty.Header>
						<Empty.Title>No architectural entities</Empty.Title>
						<Empty.Description>
							Adjust the filters or supply graph data to view the map.
						</Empty.Description>
					</Empty.Header>
				</Empty.Root>
			</div>
		{/if}
		{#if controller.error}
			<div class="absolute inset-x-3 top-16 flex justify-center">
				<Alert.Root variant="destructive" class="w-fit max-w-lg">
					<Alert.Title>Could not lay out map</Alert.Title>
					<Alert.Description>{controller.error} Current positions are retained.</Alert.Description>
					<Alert.Action>
						<Button variant="outline" size="sm" onclick={controller.retry}>Retry</Button>
					</Alert.Action>
				</Alert.Root>
			</div>
		{/if}
	</div>
</SvelteFlowProvider>
