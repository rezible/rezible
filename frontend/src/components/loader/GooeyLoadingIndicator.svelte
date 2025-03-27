<script lang="ts">
  import { Gooey } from 'svelte-ux';

</script>

<Gooey blur={4} alphaPixel={19} alphaShift={-9}>
	<div class="w-40 h-40 grid grid-stack place-items-center animate-spin [animation-duration:3s]">
		{#each { length: 6 } as _, i}
			<div
				class="orbit-converge w-6 h-6 rounded-full"
				style:--orbit-radius="60px"
				style:--orbit-start="{35 * i}deg"
				style:background-color="hsl({180 + i * 5}, 80%, 50%)"
			></div>
		{/each}
	</div>
</Gooey>
