<script lang="ts">
	import { writable } from "svelte/store";

	import {
		SvelteFlow,
		Controls,
		Background,
		BackgroundVariant,
		MiniMap,
		type SnapGrid,
		type Node,
		type Edge,
		MarkerType,
	} from "@xyflow/svelte";
	import "@xyflow/svelte/dist/style.css";

	import ContextMenu, { type ContextMenuProps } from "./ContextMenu.svelte";
    import IncidentTimeline from "./IncidentTimeline.svelte";

	const nodes = writable<Node[]>([
		{
			id: "service-1",
			data: { label: "API Service" },
			position: { x: 0, y: 0 },
		},
		{
			id: "control-1",
			data: { label: "Rate Limiter" },
			position: { x: 100, y: 100 },
		},
	]);

	const edges = writable<Edge[]>([
		{
			id: "e1",
			source: "service-1",
			target: "control-1",
			type: "control",
			label: "Controls",
			markerEnd: {
				type: MarkerType.ArrowClosed,
			},
		},
	]);

	const snapGrid: SnapGrid = [25, 25];


	let menu = $state<ContextMenuProps>();
	let width = $state(0);
	let height = $state(0);

	const handleContextMenu = (node: Node, event: MouseEvent | TouchEvent) => {
		event.preventDefault();
		const clientX = "clientX" in event ? event.clientX : 0;
		const clientY = "clientY" in event ? event.clientY : 0;

		menu = {
			id: node.id,
			top: clientY < height - 200 ? clientY : undefined,
			left: clientX < width - 200 ? clientX : undefined,
			right: clientX >= width - 200 ? width - clientX : undefined,
			bottom: clientY >= height - 200 ? height - clientY : undefined,
		};
	};

	const handlePaneClick = () => {
		menu = undefined;
	};
</script>

<div class="h-full w-full">
	<div style:height="40%">
		<IncidentTimeline />
	</div>

	<div style:height="60%" bind:clientWidth={width} bind:clientHeight={height}>
		<SvelteFlow
			{nodes}
			{edges}
			{snapGrid}
			fitView
			proOptions={{ hideAttribution: true }}
			on:nodeclick={(event) =>
				console.log("on node click", event.detail.node)}
			on:paneclick={handlePaneClick}
			on:nodecontextmenu={(e) =>
				handleContextMenu(e.detail.node, e.detail.event)}
		>
			<Background variant={BackgroundVariant.Dots} />
			{#if !!menu}
				<ContextMenu onClick={handlePaneClick} {...menu} />
			{/if}
			<Controls />
			<MiniMap />
		</SvelteFlow>
	</div>
</div>
