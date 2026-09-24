<script lang="ts">
	import { untrack } from "svelte";

	import {
		Background,
		BackgroundVariant,
		Controls,
		Panel,
		SvelteFlow,
		SvelteFlowProvider,
		type EdgeTypes,
		type NodeTypes,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

	import { mode } from "mode-watcher";

	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Empty from "$components/ui/empty";
	import { MAP_MAX_ZOOM, MAP_MIN_ZOOM } from "$features/system/lib/system-map/interaction";
	import { getDetailLayerLabel } from "$features/system/lib/system-map/category";
	import type { MapSelection } from "$features/system/lib/system-map/presentation";
	import MapConnection from "./map-connection/MapConnection.svelte";
	import MapNode from "./map-node/MapNode.svelte";
	import { SystemMapController } from "./controller.svelte";
	import ViewportApi from "./viewport-api/ViewportApi.svelte";
	import type { SystemMapHandle, SystemMapProps } from "./types";

	type Props = SystemMapProps;
	let { graph, positions, selection, highlights, onSelectionChange, onNodeMove }: Props = $props();

	const controller = new SystemMapController({
		graph: () => graph,
		positions: () => positions,
		selection: () => selection,
		highlights: () => highlights,
		onSelectionChange: (nextSelection, trigger) => onSelectionChange(nextSelection, trigger),
		onNodeMove: () => onNodeMove,
	});

	const nodeTypes: NodeTypes = { "system-map-node": MapNode };
	const edgeTypes: EdgeTypes = { "system-map-connection": MapConnection };
	const observeViewport = (element: HTMLElement) => untrack(() => controller.mount(element));

	export function fit(): ReturnType<SystemMapHandle["fit"]> {
		controller.fit();
	}

	export function recenter(): ReturnType<SystemMapHandle["recenter"]> {
		controller.recenter();
	}

	export function reveal(target: MapSelection): ReturnType<SystemMapHandle["reveal"]> {
		controller.reveal(target);
	}
</script>

<SvelteFlowProvider>
	<ViewportApi {controller} />
	<div
		class="system-map flex h-full min-h-0 w-full min-w-0 overflow-hidden"
		onkeydowncapture={controller.keydown}
	>
		<section
			{@attach observeViewport}
			class="bg-muted/20 relative min-h-0 min-w-0 flex-1"
			role="application"
			aria-label="System map canvas"
		>
			<SvelteFlow
				class="rezible-flow"
				{nodeTypes}
				{edgeTypes}
				colorMode={mode.current === "dark" ? "dark" : "light"}
				proOptions={{ hideAttribution: true }}
				minZoom={MAP_MIN_ZOOM}
				maxZoom={MAP_MAX_ZOOM}
				bind:nodes={controller.nodes}
				bind:edges={controller.edges}
				bind:viewport={controller.viewport}
				nodesDraggable={controller.nodesDraggable}
				nodesConnectable={false}
				elementsSelectable={false}
				onmove={(event, viewport) => controller.onViewportChange(viewport)}
				onpaneclick={({ event }) => controller.clearSelection(event)}
				onnodeclick={({ node, event }) => controller.selectNode(node.id, event)}
				onedgeclick={({ edge, event }) => controller.selectEdge(edge.id, event)}
				onnodedragstart={controller.onNodeDragStart}
				onnodedrag={controller.onNodeDrag}
				onnodedragstop={controller.onNodeDragStop}
				onedgepointerenter={controller.onEdgePointerEnter}
				onedgepointerleave={controller.onEdgePointerLeave}
			>
				<Background variant={BackgroundVariant.Dots} />
				<Controls showFitView={false} showLock={false} />

				<Panel position="top-left">
					<div
						class="bg-background border-border flex items-center gap-1 rounded-md border p-1 shadow-sm"
					>
						<Button variant="outline" size="sm" onclick={controller.fit}>Fit</Button>
						<Button variant="outline" size="sm" onclick={controller.recenter}>Recenter</Button>
						<Button
							variant="outline"
							size="sm"
							disabled={!controller.canReveal}
							onclick={controller.revealSelected}
						>
							Reveal
						</Button>
						<Button
							variant={controller.showAllConnectionLabels ? "secondary" : "outline"}
							size="sm"
							aria-pressed={controller.showAllConnectionLabels}
							aria-label="Show all connection labels"
							onclick={controller.toggleConnectionLabels}
						>
							Labels
						</Button>
						<Badge
							variant="secondary"
							aria-label={`${getDetailLayerLabel(controller.continuousDetail)} layer; detail ${controller.continuousDetail.toFixed(2)}`}
						>
							{getDetailLayerLabel(controller.continuousDetail)} · Detail {controller.continuousDetail.toFixed(
								2
							)}
						</Badge>
					</div>
				</Panel>

				<Panel position="bottom-left">
					{#if controller.loading}
						<div
							role="status"
							class="bg-background border-border rounded-md border px-3 py-2 text-sm shadow-sm"
						>
							Updating map…
						</div>
					{/if}
				</Panel>
			</SvelteFlow>

			{#if controller.overviewEmptyState}
				<div class="pointer-events-none absolute inset-0 flex items-center justify-center p-6">
					<Empty.Root
						role="status"
						class="pointer-events-auto max-w-sm border-border bg-background/95 shadow-sm"
					>
						<Empty.Header>
							{#if controller.overviewEmptyState === "truly-empty"}
								<Empty.Title>No graph data</Empty.Title>
								<Empty.Description>No graph entities were supplied.</Empty.Description>
							{:else}
								<Empty.Title>No architectural entities</Empty.Title>
								<Empty.Description>
									The supplied records do not include architectural nodes for the canvas.
								</Empty.Description>
							{/if}
						</Empty.Header>
					</Empty.Root>
				</div>
			{/if}

			{#if controller.error}
				<div class="pointer-events-none absolute inset-x-3 top-16 flex justify-center">
					<Alert.Root variant="destructive" class="pointer-events-auto w-fit max-w-lg shadow-sm">
						<Alert.Title>Map layout failed</Alert.Title>
						<Alert.Description>
							{controller.error}
							{#if controller.hasDiagram}
								{#if controller.showingOlderSuppliedData}
									The last valid map represents earlier supplied data.
								{:else}
									The last valid map is still shown.
								{/if}
							{/if}
						</Alert.Description>
						<Alert.Action>
							<Button variant="outline" size="sm" onclick={controller.retry}>Retry</Button>
						</Alert.Action>
					</Alert.Root>
				</div>
			{/if}
		</section>
	</div>
</SvelteFlowProvider>

<style>
	:global(.system-map .svelte-flow__edge-path) {
		transition: stroke-width 180ms ease;
	}

	:global(.system-map .svelte-flow__edge-label) {
		pointer-events: none !important;
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.system-map .svelte-flow__edge-path) {
			transition: none;
		}
	}
</style>
