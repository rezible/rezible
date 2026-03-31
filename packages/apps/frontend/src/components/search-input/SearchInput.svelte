<script lang="ts">
	import { mdiMagnify } from "@mdi/js";
	import type { ComponentProps } from "svelte";

	type Props = {
		value: string | undefined;
	};
	let {
		value = $bindable(),
	}: Props = $props();

	const coerceValue = (v: string | number | null) => ((!!v && typeof v === "string") ? v : undefined);
</script>

<span>search field</span>
<!--TextField
	label="Search"
	labelPlacement="top"
	on:change={(e) => (value = coerceValue(e.detail.inputValue))}
	debounceChange={300}
	iconRight={mdiMagnify}
	...fieldProps
/-->