<script lang="ts">
	import { createEventDispatcher, type ComponentProps, type Snippet } from "svelte";
	import { Button } from "svelte-ux";

	type Props = ComponentProps<Button> & {
		// onclick?: (e: MouseEvent) => void;
		children?: Snippet;
	};
	const props: Props = $props();

	const dispatch = createEventDispatcher();
	const onClicked = (e: MouseEvent) => {
		// if (props.onclick) {
		// 	props.onclick(e);
		// } else {
			dispatch("click", e);
		// }
	}
</script>

<Button {...props} on:click={onClicked}>
	{@render props.children?.()}
</Button>