<script lang="ts">
	import { mdiBug, mdiPlusNetwork } from '@mdi/js';
    import { useNodes } from '@xyflow/svelte';
	import { Button, Icon } from 'svelte-ux';

	const nodes = useNodes();
</script>

<div class="absolute bottom-2 w-full">
	<div class="w-fit group flex gap-2 justify-center mx-auto p-2 rounded-lg border border-surface bg-surface-100/20">
		<Button color="primary" variant="fill">
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">Add Node</span>
				<Icon data={mdiPlusNetwork} />
			</span>
		</Button>
		<Button color="default" variant="fill" on:click={() => {console.log($nodes)}}>
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">log</span>
				<Icon data={mdiBug} />
			</span>
		</Button>
	</div>
</div>