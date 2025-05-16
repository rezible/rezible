<script lang="ts">
	import { useConnection } from "@xyflow/svelte";

	const connection = useConnection();
	
	const from = $derived(connection.current.from || {x: 0, y: 0});
	const to = $derived(connection.current.to || {x: 0, y: 0});
	const path = $derived(`M${from.x},${from.y} C ${from.x} ${to.y} ${from.x} ${to.y} ${to.x},${to.y}`);
	const fromHandle = $derived(connection.current.fromHandle);
</script>

{#if path}
	<path
		fill="none"
		stroke-width={1.5}
		class="animated stroke-primary"
		stroke={fromHandle?.id}
		d={path}
	/>
	<circle cx={to.x} cy={to.y} fill="#fff" r={3} stroke={fromHandle?.id} stroke-width={1.5} />
{/if}
