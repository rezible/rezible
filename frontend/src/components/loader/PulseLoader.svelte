<script lang="ts">
	let { 
		color = '#FF3E00',
		unit = 'px',
		duration = 1.5,
		size = 42,
		pause = false,
	 } = $props();
</script>

<div class="wrapper" style="--size: {size}{unit}; --color: {color}; --duration: {duration}s">
	{#each [2, 1, 0] as version}
		<div
			class="cube bg-primary-600"
			class:pause-animation={pause}
			style="animation-delay: {version * (duration / 10)}s; left: {version *
				(size / 3 + size / 15) +
				unit};"
		></div>
	{/each}
</div>

<style>
	.wrapper {
		position: relative;
		display: flex;
		justify-content: center;
		align-items: center;
		width: var(--size);
		height: calc(var(--size) / 2.5);
	}
	.cube {
		position: absolute;
		top: 0px;
		width: calc(var(--size) / 5);
		height: calc(var(--size) / 2.5);
		/*background-color: var(--color);*/
		animation: motion var(--duration) cubic-bezier(0.895, 0.03, 0.685, 0.22) infinite;
	}
	.pause-animation {
		animation-play-state: paused;
	}
	@keyframes motion {
		0% {
			opacity: 1;
		}
		50% {
			opacity: 0;
		}
		100% {
			opacity: 1;
		}
	}
</style>
