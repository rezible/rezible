<script lang="ts">
	import { ElementSize } from "runed";
	import { Button } from "svelte-ux";

	type Props = {
		containerRect: DOMRect;
		clickPos: {x: number; y: number};
		timestamp: number;
		close: () => void;
	};
	const {containerRect, clickPos, timestamp, close}: Props = $props();

	const onClicked = (e: MouseEvent) => {
		e.stopPropagation();
	};

	let containerRef = $state<HTMLDivElement>(null!);
	const size = new ElementSize(() => containerRef);

	const xOverflows = $derived((clickPos.x + size.width) > containerRect.right);
	const naiveX = $derived(Math.round(clickPos.x - containerRect.x));
	const posX = $derived(xOverflows ? (naiveX - size.width) : naiveX);

	const yOverflows = $derived((clickPos.y + size.height) > containerRect.bottom);
	const naiveY = $derived(Math.round(clickPos.y - containerRect.y));
	const posY = $derived(yOverflows ? (naiveY - size.height) : naiveY);
</script>

<div 
	id="timeline-ctx-container"
	class="w-48 border bg-surface-300 w-full flex flex-col gap-2 p-2 h-fit"
	style="position: absolute; left: {posX}px; top: {posY}px"
	role="presentation"
	onclick={onClicked}
	bind:this={containerRef}
>
	<span>{new Date(timestamp).toString()}</span>
	<Button color="accent" variant="fill" on:click={() => {}}>Add Event</Button>
</div>