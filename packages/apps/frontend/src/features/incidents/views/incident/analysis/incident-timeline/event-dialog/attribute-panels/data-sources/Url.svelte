<script lang="ts">
	import { Button } from "$components/ui/button";
	import RiLink from "remixicon-svelte/icons/link";
	import RiAddLine from "remixicon-svelte/icons/add-line";

	type Props = {
		dataValue: string;
	};
	let { dataValue = $bindable() }: Props = $props();
</script>

<!-- <TextField label="Link New Item" bind:value={dataValue} classes={{ input: "text-sm" }}>
	<div slot="prepend">
		<RiLink class="text-surface-content/50 mr-2" aria-hidden="true" />
	</div>
	<span slot="append">
		<Button class="text-surface-content/50 p-2" />
	</span>
</TextField> -->
