<script lang="ts">
	import { mdiBug, mdiPlusNetwork } from "@mdi/js";
	import { useNodes } from "@xyflow/svelte";
	import { Button, Icon } from "svelte-ux";
	import { analysis } from "../analysis.svelte";

	type Props = {};
	let {}: Props = $props();

	const nodes = useNodes();
</script>

<div class="absolute bottom-2 right-2 w-fit mx-auto">
	<div
		class="w-fit group flex gap-2 justify-center mx-auto p-2 rounded-lg border border-surface bg-surface-100/20"
	>
		<Button
			color="primary"
			variant="fill"
			on:click={() => analysis.setComponentDialogOpen(true)}
		>
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">Add Component</span>
				<Icon data={mdiPlusNetwork} />
			</span>
		</Button>
		<Button
			color="default"
			variant="fill"
			on:click={() => {
				console.log($nodes);
			}}
		>
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">log</span>
				<Icon data={mdiBug} />
			</span>
		</Button>
	</div>
</div>
