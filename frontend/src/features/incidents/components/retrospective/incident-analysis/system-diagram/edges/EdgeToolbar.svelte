<script lang="ts">
    // import { useNodeConnections } from "@xyflow/svelte";
    import { diagram } from "../diagram.svelte";
    import { watch } from "runed";
    import { get } from "svelte/store";

	const edge = $derived(diagram.selectedEdge);
	
	watch(() => diagram.selectedEdge, (e) => {
		if (!e) return;
		// get()
		
	})
</script>

{#if edge}
	<span>selected: {edge.data}</span>
{/if}