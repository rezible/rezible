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
		if (!props.onclick) {
			alert("button using legacy event handler");
			dispatch("click", e);
		}
		props.onclick?.(e);
	}
</script>

<Button {...props} on:click={onClicked}>
	{@render props.children?.()}
</Button>