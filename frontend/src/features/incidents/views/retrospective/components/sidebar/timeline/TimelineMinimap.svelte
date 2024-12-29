 <script lang="ts">
	import { onMount } from "svelte";

	type Props = {
		containerEl: HTMLElement;
		events: any[];
		onEventClicked: (id: string) => void;
	}
	const { containerEl, events, onEventClicked }: Props = $props();

	let progressBarContainerEl = $state<HTMLElement>();

	let scrollHeight = $state(containerEl.scrollHeight);
	let clientHeight = $state(containerEl.clientHeight);
	let progessBarContainerHeight = $state(containerEl.clientHeight);

	const windowHeight = $derived(progessBarContainerHeight * (clientHeight / scrollHeight));

	const updateHeights = () => {
		scrollHeight = containerEl.scrollHeight;
		clientHeight = containerEl.clientHeight;
		progessBarContainerHeight = progressBarContainerEl?.scrollHeight || clientHeight;
	}
	
	let scrollTop = $state(containerEl.scrollTop);
	const updateScrollTop = () => {scrollTop = containerEl.scrollTop};
	
	const scrollAmount = $derived(100 * ((clientHeight + scrollTop) / scrollHeight));

	const observer = new ResizeObserver(updateHeights);
	onMount(() => {
		updateHeights();
		observer.observe(containerEl);
		containerEl.addEventListener("scroll", updateScrollTop);
		return () => {
			containerEl.removeEventListener("scroll", updateScrollTop);
			observer.disconnect();
		}
	});
 </script>
 
<div class="w-8 flex flex-col justify-evenly bg-surface-content/25 overflow-y-clip" 
	style="--scrollAmount: {scrollAmount}%; --windowHeight: {windowHeight}px" 
	bind:this={progressBarContainerEl}
>
	<div id="progress-bar-filled"></div>
	<div id="progress-bar"></div>
</div>

<style>
	#progress-bar-filled {
		display: flex;
		flex: var(--scrollAmount);
	}

	#progress-bar-filled::after {
		width: 2rem;
		height: var(--windowHeight);
		align-self: flex-end;
		background-color: oklch(var(--color-surface-content) / 0.25);
		content: var(--tw-content);
	}

	#progress-bar {
		width: 2rem;
		flex: calc(100% - var(--scrollAmount));
	}
</style>