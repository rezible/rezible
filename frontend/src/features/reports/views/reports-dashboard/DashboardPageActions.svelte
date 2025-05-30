<script lang="ts">
	import { mdiPlus } from "@mdi/js";
	import { Button } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
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
