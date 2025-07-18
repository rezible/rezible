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

	const coerceValue = (v: string | number | null) => ((!!v && typeof v === "string") ? v : undefined);
</script>

<TextField
	label="Search"
	labelPlacement="top"
	on:change={(e) => (value = coerceValue(e.detail.inputValue))}
	debounceChange={300}
	iconRight={mdiMagnify}
	...fieldProps
/>