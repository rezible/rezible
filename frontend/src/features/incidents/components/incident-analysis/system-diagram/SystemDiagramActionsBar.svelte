<script lang="ts">
	import { mdiShapeSquareRoundedPlus } from "@mdi/js";
	import { useNodes } from "@xyflow/svelte";
	import { Button, Icon } from "svelte-ux";
	import { componentDialog } from "$features/incidents/components/incident-analysis/system-diagram/component-dialog/dialogState.svelte";

	type Props = {};
	let {}: Props = $props();

	const nodes = useNodes();
</script>

<div class="absolute bottom-2 right-2 w-fit mx-auto">
	<div
		class="w-fit group flex flex-col gap-2 justify-center mx-auto p-1 rounded-lg border border-surface bg-surface-100/20"
	>
		<Button color="primary" variant="fill" on:click={componentDialog.setAdding}>
			<span class="flex items-center gap-2">
				<span class="group-hover:inline hidden">Add Component</span>
				<Icon data={mdiShapeSquareRoundedPlus} />
			</span>
		</Button>
	</div>
</div>
