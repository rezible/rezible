<script lang="ts">
    import { ViewportPortal, useSvelteFlow } from "@xyflow/svelte";
    import { diagram } from "./diagram.svelte";
    import { Button, ButtonGroup } from "svelte-ux";

	// const {viewport} = useSvelteFlow();
	// let zoom = $state($viewport.zoom);
	// viewport.subscribe(v => zoom = v.zoom);

	const {node, edge} = $derived(diagram.selected);
	const visible = $derived(!!node || !!edge);

	//const containerTransformStyle = $derived((pos2 && shift) ? `transform: translate(${pos2[0]}px, ${pos2[1]}px) translate(${shift[0]}%, ${shift[1]}%);` : "");
	const containerTransformStyle = $derived(visible ? `transform: translate(-50%, -50%) translate(${diagram.toolbarPosition.x}px, ${diagram.toolbarPosition.y}px);` : "");
</script>

<ViewportPortal>
	{#if visible}
		<div class="pointer-events-auto absolute border rounded-lg bg-surface-100 p-1 z-10 svelte-flow__node-toolbar" style={containerTransformStyle}>
			<ButtonGroup variant="fill-light" color="accent" size="sm">
				{#if node}
					<Button on:click={() => console.log("cmp")}>Edit Component</Button>
				{:else if edge}
					<Button on:click={() => console.log("rel")}>Edit Relationship</Button>
				{/if}
			</ButtonGroup>
		</div>
	{/if}
</ViewportPortal>