<script lang="ts">
	let { 
		durationMs = 2000,
		size = 48,
		pause = false,
	 } = $props();
</script>

<div class="wrapper" style="--size: {size}px; --duration: {durationMs}ms;">
	<div class="rez-loading-ring ring-one" class:pause-animation={pause}></div>
	<div class="rez-loading-ring ring-two" class:pause-animation={pause}></div>
</div>

<style lang="postcss">
	.wrapper {
		position: relative;
		width: var(--size);
		height: var(--size);
	}
	.rez-loading-ring {
		@apply border-accent-300;
		position: absolute;
		top: 0px;
		left: 0px;
		width: var(--size);
		height: var(--size);
		opacity: 0.4;
		perspective: 800px;
		border-width: 6px;
		border-style: solid;
		border-image: initial;
		border-radius: 100%;
	}
	.rez-loading-ring.ring-one {
		animation: var(--duration) linear 0s infinite normal none running ringOne;
	}
	.rez-loading-ring.ring-two {
		animation: var(--duration) linear 0s infinite normal none running ringTwo;
	}
	.pause-animation {
		animation-play-state: paused;
	}

	@keyframes ringOne {
		0% {
			transform: rotateX(0deg) rotateY(0deg) rotateZ(0deg);
		}
		100% {
			transform: rotateX(360deg) rotateY(180deg) rotateZ(360deg);
		}
	}
	@keyframes ringTwo {
		0% {
			transform: rotateX(0deg) rotateY(0deg) rotateZ(0deg);
		}
		100% {
			transform: rotateX(180deg) rotateY(360deg) rotateZ(360deg);
		}
	}
</style>