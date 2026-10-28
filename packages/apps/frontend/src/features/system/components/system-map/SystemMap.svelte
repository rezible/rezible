<script lang="ts">
	import {
		Controls,
		Panel,
		SvelteFlow,
		SvelteFlowProvider,
		type NodeTypes,
		type EdgeTypes,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

	import { mode } from "mode-watcher";
	import * as NativeSelect from "$components/ui/native-select";
	import type { RelationshipMode } from "$features/system/lib/system-map/rollup";
	import type { DetailMode } from "$features/system/lib/system-map/detail";
	import * as Empty from "$components/ui/empty";

	import type { MapHighlights, MapSelection } from "$features/system/lib/system-map/presentation";
	import type { GraphSubset } from "$features/system/lib/system-map/graph";

	import MapCard from "./map-card/MapCard.svelte";
	import MapConnection from "./map-connection/MapConnection.svelte";
	import MapRegion from "./map-region/MapRegion.svelte";
	import MapRegionLabels from "./map-region-labels/MapRegionLabels.svelte";
	import SystemMapViewport from "./SystemMapViewport.svelte";
	import { setSystemMapController, SystemMapController } from "./controller.svelte";

	export type SystemMapProps = {
		graph: GraphSubset;
		selection?: MapSelection;
		highlights?: MapHighlights;
		/** Emits the user's selection intention without owning an independent selection. */
		onSelectionChange: (selection: MapSelection | undefined, trigger?: HTMLElement) => void;
		/** Opens the host's deployment view at a placement target; without it, cards do not offer to. */
		onExploreDeployment?: (targetId: string) => void;
	};

	const { graph, selection, highlights, onSelectionChange, onExploreDeployment }: SystemMapProps = $props();

	const controller = setSystemMapController(
		new SystemMapController({
			graph: () => graph,
			selection: () => selection,
			highlights: () => highlights,
			onSelectionChange: (next, trigger) => onSelectionChange(next, trigger),
			onExploreDeployment: () => onExploreDeployment,
		})
	);
	const detailId = $props.id();
	const relationshipsId = `${detailId}-relationships`;
	const nodeTypes: NodeTypes = { "system-map-region": MapRegion, "system-map-card": MapCard };
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
	<div
		class="system-map relative h-full min-h-0 w-full"
		style:--map-zoom={controller.zoom}
		onkeydowncapture={controller.keydown}
	>
		<!-- Manual z-index keeps the tiers fixed: regions, connections, cards, then landscape band labels. -->
		<SvelteFlow
			{nodeTypes}
			{edgeTypes}
			bind:nodes={controller.nodes}
			bind:edges={controller.edges}
			colorMode={mode.current === "dark" ? "dark" : "light"}
			proOptions={{ hideAttribution: true }}
			maxZoom={2}
			zIndexMode="manual"
			nodesDraggable={false}
			nodesConnectable={false}
			elementsSelectable={false}
			deleteKey={null}
			onnodeclick={({ node, event }) => controller.selectNode(node.id, event)}
			onedgeclick={({ edge, event }) => controller.selectEdge(edge.id, event)}
			onedgepointerenter={({ edge }) => controller.hoverEdge(edge.id)}
			onedgepointerleave={() => controller.hoverEdge(undefined)}
			onpaneclick={controller.clearSelection}
			onmove={(_, viewport) => controller.viewportZoom(viewport.zoom)}
		>
			<Panel position="top-left" class="bg-card border-border space-y-2 rounded-md border p-2">
				<div class="flex items-center gap-2 text-sm">
					<label for={detailId}>Detail</label>
					<NativeSelect.Root
						id={detailId}
						size="sm"
						value={controller.detailMode}
						onchange={(event) =>
							controller.setDetailMode(event.currentTarget.value as DetailMode)}
					>
						<NativeSelect.Option value="auto">Auto</NativeSelect.Option>
						<NativeSelect.Option value="landscape">Landscape</NativeSelect.Option>
						<NativeSelect.Option value="runtime">Runtime</NativeSelect.Option>
					</NativeSelect.Root>
				</div>
				<div class="flex items-center gap-2 text-sm">
					<label for={relationshipsId}>Relationships</label>
					<NativeSelect.Root
						id={relationshipsId}
						size="sm"
						value={controller.relationshipMode}
						onchange={(event) =>
							controller.setRelationshipMode(event.currentTarget.value as RelationshipMode)}
					>
						<NativeSelect.Option value="focused">Focused</NativeSelect.Option>
						<NativeSelect.Option value="all">All</NativeSelect.Option>
					</NativeSelect.Root>
				</div>
			</Panel>
			<Panel
				position="bottom-center"
				class="bg-card border-border text-muted-foreground flex items-center gap-4 rounded-md border px-3 py-2 text-xs"
			>
				<span class="flex items-center gap-1.5">
					<span class="swatch landscape" aria-hidden="true"></span>
					Landscape
				</span>
				<span class="flex items-center gap-1.5">
					<span class="swatch system" aria-hidden="true"></span>
					System
				</span>
				<span class="flex items-center gap-1.5">
					<span class="swatch runtime" aria-hidden="true"></span>
					Runtime
				</span>
				<span class="border-border border-l pl-4">Implementation grouped</span>
			</Panel>
			<SystemMapViewport {controller} />
			{#if controller.band === "landscape"}
				<MapRegionLabels nodes={controller.nodes} />
			{/if}
			<Controls showFitView={false} showLock={false} />
		</SvelteFlow>
		{#if !controller.nodes.length}
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
	</div>
</SvelteFlowProvider>

<style>
	.system-map :global(.svelte-flow) {
		--xy-background-color: var(--background);
	}

	.system-map :global(.svelte-flow__node:focus-visible) {
		outline: none;
	}

	/* The tier hues at the region fill and stroke strengths. */
	.swatch {
		width: 10px;
		height: 10px;
		border-radius: 9999px;
		background: color-mix(in srgb, var(--tier) 25%, var(--card));
		border: 1px solid color-mix(in srgb, var(--tier) 60%, transparent);
	}

	.swatch.landscape {
		--tier: var(--chart-1);
	}

	.swatch.system {
		--tier: var(--chart-4);
	}

	.swatch.runtime {
		--tier: var(--chart-2);
	}
</style>
