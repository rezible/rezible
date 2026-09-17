<script lang="ts">
	import { onMount } from "svelte";
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

	import type { GraphSubset } from "$features/systems/lib/system-map/graph";
	import type { MapDisplayOptions } from "$features/systems/lib/system-map/presentation";
	import type { Viewport } from "$features/systems/lib/system-map/geometry";
	import { MAP_MAX_ZOOM, MAP_MIN_ZOOM } from "$features/systems/lib/system-map/interaction";
	import MapConnection from "./map-connection/MapConnection.svelte";
	import MapInspector from "./map-inspector/MapInspector.svelte";
	import MapNode from "./map-node/MapNode.svelte";

	import { initSystemMapController, type SystemMapSelection } from "./controller.svelte";

	type Props = {
		graph: GraphSubset;
		displayOptions: MapDisplayOptions;
		onSelect?: (selection: SystemMapSelection) => void;
		onClearSelection?: () => void;
		onViewportChange?: (viewport: Viewport) => void;
	};

	let { graph, displayOptions, onSelect, onClearSelection, onViewportChange }: Props = $props();

	const controller = initSystemMapController({
		graph: () => graph,
		displayOptions: () => displayOptions,
		onSelect: (selection) => onSelect?.(selection),
		onClearSelection: () => onClearSelection?.(),
		onViewportChange: (viewport) => onViewportChange?.(viewport),
	});

	let inspectorHiddenAtSelectionRevision = $state<number>();
	const inspectorVisible = $derived(
		controller.selectionRevision !== inspectorHiddenAtSelectionRevision
	);

	const nodeTypes: NodeTypes = { "system-map-node": MapNode };
	const edgeTypes: EdgeTypes = { "system-map-connection": MapConnection };

	let viewportEl = $state<HTMLElement>(null!);
	onMount(() => controller.mount(viewportEl));
</script>

<SvelteFlowProvider>
	<div
		class="system-map flex h-full min-h-0 w-full min-w-0 overflow-hidden"
		onkeydowncapture={controller.keydown}
	>
		<section
			bind:this={viewportEl}
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
				nodesDraggable={false}
				nodesConnectable={false}
				elementsSelectable={false}
				onnodeclick={({ node, event }) => controller.selectNode(node.id, event)}
				onedgeclick={({ edge, event }) => controller.selectEdge(edge.id, event)}
				onpaneclick={() => controller.clearSelection()}
				onmove={controller.onMove}
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
							disabled={!controller.selectedTarget}
							onclick={controller.revealSelected}
						>
							Reveal
						</Button>
						<Badge
							variant="secondary"
							aria-label={`Detail ${controller.continuousDetail.toFixed(2)}`}
						>
							Detail {controller.continuousDetail.toFixed(2)}
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

				{#if !inspectorVisible}
					<Panel position="top-right">
						<Button
							variant="outline"
							size="sm"
							onclick={() => (inspectorHiddenAtSelectionRevision = undefined)}
						>
							Show inspector
						</Button>
					</Panel>
				{/if}
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
								<Empty.Description>The supplied entities are available in the source list.</Empty.Description>
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
								The last valid diagram is still shown.
							{/if}
						</Alert.Description>
						<Alert.Action>
							<Button variant="outline" size="sm" onclick={controller.retry}>Retry</Button>
						</Alert.Action>
					</Alert.Root>
				</div>
			{/if}
		</section>

		{#if inspectorVisible}
			<MapInspector
				onHide={() => (inspectorHiddenAtSelectionRevision = controller.selectionRevision)}
			/>
		{/if}
	</div>
</SvelteFlowProvider>

<style>
	:global(.system-map .svelte-flow__edge-path) {
		transition: stroke-width 180ms ease;
	}

	:global(.system-map .svelte-flow__edge-label) {
		/* EdgeLabel sets inline pointer-events: all; keep labels presentation-only. */
		/* Hits then reach only the underlying edge hit region. */
		pointer-events: none !important;
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.system-map .svelte-flow__edge-path) {
			transition: none;
		}
	}
</style>
