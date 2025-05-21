<script lang="ts">
	import { mdiShapeSquareRoundedPlus } from "@mdi/js";
	import { Button, Icon } from "svelte-ux";
	import { useComponentDialog } from "./component-dialog/dialogState.svelte";

	const componentDialog = useComponentDialog();
</script>

<div
	class="w-fit group flex flex-col gap-2 justify-center mx-auto p-1 rounded-lg border border-surface bg-surface-100/20"
>
	<Button color="primary" variant="fill" on:click={() => componentDialog.setAdding()}>
		<span class="flex items-center gap-2">
			<span class="group-hover:inline hidden">Add Component</span>
			<Icon data={mdiShapeSquareRoundedPlus} />
		</span>
	</Button>
</div>