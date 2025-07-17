<script lang="ts">
	import { mdiMagnify } from "@mdi/js";
	import type { ComponentProps } from "svelte";
	import { TextField } from "svelte-ux";

	type Props = {
		value: string | undefined;
		field?: ComponentProps<TextField>;
	};
	let {
		value = $bindable(),
		field: fieldProps = {},
	}: Props = $props();
</script>

<TextField
	label="Search"
	labelPlacement="top"
	on:change={(e) => console.log(e.detail)}
	debounceChange
	iconRight={mdiMagnify}
	...fieldProps
/>