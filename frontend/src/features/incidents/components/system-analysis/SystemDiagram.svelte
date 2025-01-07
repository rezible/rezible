<script lang="ts">
	import {
		SvelteFlow,
		Controls,
		Background,
		BackgroundVariant,
		MiniMap,
	} from "@xyflow/svelte";

	import "@xyflow/svelte/dist/style.css";

	import ContextMenu from "./SystemDiagramContextMenu.svelte";
    import { diagram } from "./diagram.svelte";

	let containerEl = $state<HTMLElement>();
	diagram.componentSetup(() => containerEl);
</script>

<div class="h-full w-full" bind:this={containerEl}>
	<SvelteFlow
		nodes={diagram.nodes}
		edges={diagram.edges}
		snapGrid={[25, 25]}
		fitView
		proOptions={{ hideAttribution: true }}
		on:nodeclick={e => diagram.handleNodeClicked(e.detail)}
		on:paneclick={e => diagram.handlePaneClicked(e.detail.event)}
		on:nodecontextmenu={e => diagram.handleNodeContextMenu(e.detail)}
	>
		<Background variant={BackgroundVariant.Dots} />
		{#if !!diagram.ctxMenuProps}
			<ContextMenu {...diagram.ctxMenuProps} />
		{/if}
		<Controls />
		<MiniMap />
	</SvelteFlow>
</div>
