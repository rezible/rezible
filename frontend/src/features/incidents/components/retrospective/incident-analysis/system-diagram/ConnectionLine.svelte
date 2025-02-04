<script lang="ts">
	import { useConnection, type XYPosition } from "@xyflow/svelte";

	const connection = useConnection();

	let from = $state<XYPosition>({ x: 0, y: 0 });
	let to = $state<XYPosition>({ x: 0, y: 0 });
	let path = $state("");

	connection.subscribe((c) => {
		if (c.from && c.to) {
			from = c.from;
			to = c.to;
			path = `M${from.x},${from.y} C ${from.x} ${to.y} ${from.x} ${to.y} ${to.x},${to.y}`;
			/*
			[path] = getSmoothStepPath({
				sourcePosition: c.fromPosition,
				targetPosition: c.toPosition,
				targetX: to.x,
				targetY: to.y,
				sourceX: from.x,
				sourceY: from.y,
			});
			*/
		}
	});
</script>

{#if path}
	<path
		fill="none"
		stroke-width={1.5}
		class="animated stroke-primary"
		stroke={$connection.fromHandle?.id}
		d={path}
	/>
	<circle
		cx={to.x}
		cy={to.y}
		fill="#fff"
		r={3}
		stroke={$connection.fromHandle?.id}
		stroke-width={1.5}
	/>
{/if}
