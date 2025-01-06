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
	} from "@xyflow/svelte";

	import "@xyflow/svelte/dist/style.css";

	import ContextMenu, { type ContextMenuProps } from "./ContextMenu.svelte";
    import { systemView } from "../analysis.svelte";

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

<div class="h-full w-full" bind:clientWidth={width} bind:clientHeight={height}>
	<SvelteFlow
		nodes={systemView.nodes}
		edges={systemView.edges}
		snapGrid={[25, 25]}
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
