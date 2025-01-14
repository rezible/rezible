<script lang="ts">
	import { mdiPlusNetwork } from '@mdi/js';
	import { Button, Icon, Tooltip } from 'svelte-ux';
</script>

<div class="absolute bottom-2 w-full">
	<div class="w-fit group flex gap-2 justify-center mx-auto px-3 py-1 rounded-lg border border-surface bg-surface-100/20">
		<Button color="primary" variant="fill">
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">Add Node</span>
				<Icon data={mdiPlusNetwork} />
			</span>
		</Button>
	</div>
</div>