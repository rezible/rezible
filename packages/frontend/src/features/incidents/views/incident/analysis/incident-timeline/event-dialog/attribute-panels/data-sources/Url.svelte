<script lang="ts">
	import { Button } from "$components/ui/button";
	import { mdiLink, mdiPlus } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";

	type Props = {
		dataValue: string;
	};
	let { dataValue = $bindable() }: Props = $props();
</script>

<!-- <TextField label="Link New Item" bind:value={dataValue} classes={{ input: "text-sm" }}>
	<div slot="prepend">
		<Icon data={mdiLink} classes={{root: "text-surface-content/50 mr-2"}} />
	</div>
	<span slot="append">
		<Button class="text-surface-content/50 p-2" />
	</span>
</TextField> -->
