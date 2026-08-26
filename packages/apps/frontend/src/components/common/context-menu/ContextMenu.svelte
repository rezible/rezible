<script lang="ts">
	import Header from "$components/layout/header/Header.svelte";
	import type { XYPosition } from "@xyflow/svelte";
	import { ElementSize } from "runed";
	import type { Snippet } from "svelte";

	type Props = {
		title: string;
		containerRect: DOMRect;
		clickPos: XYPosition;
		children: Snippet;
	};
	const { title, containerRect, clickPos, children }: Props = $props();

	const makeBoundedPositionForClick = (container: DOMRect, size: ElementSize, click: XYPosition) => {
		const xOverflows = click.x + size.width > container.right;
		const naiveX = Math.round(click.x - container.x);
		const left = xOverflows ? naiveX - size.width : naiveX;

		const yOverflows = click.y + size.height > container.bottom;
		const naiveY = Math.round(click.y - container.y);
		const top = yOverflows ? naiveY - size.height : naiveY;

		return { left, top };
	}

	let ref = $state<HTMLElement>(null!);
	const refSize = new ElementSize(() => ref);

	const pos = $derived(makeBoundedPositionForClick(containerRect, refSize, clickPos));
</script>

<div
	style="left: {pos.left}px; top: {pos.top}px;"
	class="absolute z-10 h-fit w-48 border border-border bg-card shadow"
	bind:this={ref}
>
	<Header {title} classes={{ root: "px-2 py-1" }} />

	{@render children()}
</div>
