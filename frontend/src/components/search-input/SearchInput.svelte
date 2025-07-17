<script lang="ts">
	import { mdiMagnify } from "@mdi/js";
	import { TextField } from "svelte-ux";

	type Props = {
		value: string | undefined;
	};
	let {
		value = $bindable(),
	}: Props = $props();
</script>

<TextField
	label="Search"
	dense
	on:change={(e) => console.log(e.detail)}
	debounceChange
	iconRight={mdiMagnify}
	labelPlacement="float"
/>