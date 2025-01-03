<script lang="ts">
	import { onMount } from "svelte";

	type Props = {
		containerEl: HTMLElement;
		eventsEl: HTMLElement;
		events: any[];
		onEventClicked: (id: string) => void;
	}
	const { containerEl, eventsEl, events, onEventClicked }: Props = $props();

	let scrollHeight = $state(eventsEl.scrollHeight);
	let clientHeight = $state(eventsEl.clientHeight);
	let containerHeight = $state(containerEl.clientHeight);

	const updateHeights = () => {
		scrollHeight = eventsEl.scrollHeight;
		clientHeight = eventsEl.clientHeight;
		containerHeight = containerEl.scrollHeight;
	}

	let scrollTop = $state(eventsEl.scrollTop);
	const updateScrollTop = () => {scrollTop = eventsEl.scrollTop};
	const observer = new ResizeObserver(updateHeights);
	onMount(() => {
		updateHeights();
		observer.observe(containerEl);
		eventsEl.addEventListener("scroll", updateScrollTop);
		return () => {
			eventsEl.removeEventListener("scroll", updateScrollTop);
			observer.disconnect();
		}
	});

	const windowHeight = $derived(containerHeight * (clientHeight / scrollHeight));
	const scrollAmount = $derived(100 * ((clientHeight + scrollTop) / scrollHeight));
</script>

<div class="w-8 h-full flex flex-col justify-evenly bg-surface-content/25 overflow-y-clip" 
	style="--scrollAmount: {scrollAmount}%; --windowHeight: {windowHeight}px" 
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
		max-height: 100%;
		align-self: flex-end;
		background-color: oklch(var(--color-surface-content) / 0.25);
		content: var(--tw-content);
	}

	#progress-bar {
		width: 2rem;
		flex: calc(100% - var(--scrollAmount));
	}
</style>