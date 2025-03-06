<script lang="ts">
	import { mdiPlus } from "@mdi/js";
	import { Button, Icon } from "svelte-ux";
	let createOpen = $state(false);
</script>

<div class="flex gap-2">
	<Button
		classes={{ root: "w-fit h-fit" }}
		variant="fill"
		color="primary"
		on:click={() => {createOpen = true}}
	>
		Create New Report
		<Icon data={mdiPlus} />
	</Button>
</div>
