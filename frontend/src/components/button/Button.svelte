<script lang="ts">
	import type { ComponentProps, Snippet } from "svelte";
	import { Button } from "svelte-ux";

	type Props = ComponentProps<Button> & {
		onclick?: (e: MouseEvent) => void;
		children?: Snippet;
	};
	const props: Props = $props();
</script>

<Button {...props} on:click={e => {props.onclick?.(e)}}>
	{@render props.children?.()}
</Button>