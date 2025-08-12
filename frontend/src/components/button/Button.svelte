<script lang="ts">
	import { createEventDispatcher, type ComponentProps, type Snippet } from "svelte";
	import { Button } from "svelte-ux";

	type Props = ComponentProps<Button> & {
		onclick?: (e: MouseEvent) => void;
		children?: Snippet;
	};
	const props: Props = $props();

	const dispatch = createEventDispatcher();
	const onClicked = (e: MouseEvent) => {
		props.onclick?.(e);
		dispatch("click", e);
	}
</script>

<Button {...props} on:click={onClicked}>
	{@render props.children?.()}
</Button>