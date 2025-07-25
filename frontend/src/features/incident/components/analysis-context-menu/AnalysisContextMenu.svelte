<script lang="ts">
	import Header from "$components/header/Header.svelte";
	import { ElementSize } from "runed";
	import type { Snippet } from "svelte";


	type Props = {
		title: string;
		containerRect: DOMRect;
		clickPos: {x: number, y: number};
		children: Snippet;
	}
	const { title, containerRect, clickPos, children }: Props = $props();

	let ref = $state<HTMLElement>(null!);
	const size = new ElementSize(() => ref);

	const pos = $derived.by(() => {
		const xOverflows = (clickPos.x + size.width) > containerRect.right;
		const naiveX = Math.round(clickPos.x - containerRect.x);

		const yOverflows = (clickPos.y + size.height) > containerRect.bottom;
		const naiveY = Math.round(clickPos.y - containerRect.y);
		
		return {
			left: xOverflows ? (naiveX - size.width) : naiveX,
			top: yOverflows ? (naiveY - size.height) : naiveY,
		};
	});
</script>


<div
	style="left: {pos.left}px; top: {pos.top}px;"
	class="absolute context-menu border bg-surface-200 w-48 h-fit"
	bind:this={ref}
>
	<Header {title} classes={{root: "px-2 py-1"}} />

	{@render children()}
</div>

<style>
	.context-menu {
		border-style: solid;
		box-shadow: 10px 19px 20px rgba(0, 0, 0, 10%);

		z-index: 10;
	}
</style>