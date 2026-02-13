<script lang="ts">
	import type { Snippet } from "svelte";
	import { cn } from "$lib/utils";
	import type { DOMAttributes } from "svelte/elements";

	type Classes = {
		root?: string;
		headerContainer?: string;
		content?: string;
		actions?: string;
	};

	type Props = DOMAttributes<any> & {
		header: Snippet;
		children?: Snippet;
		contents?: Snippet;
		actions?: Snippet;
		classes?: Classes;
	};
	const { header, children, contents, actions, classes = {}, ...restProps }: Props = $props();
</script>

<div
	{...restProps}
	class={cn(
		"relative z-0 bg-surface-100 border rounded elevation-1 flex flex-col justify-between",
		classes.root
	)}
>
	<div class={cn("p-4", classes.headerContainer)}>
		{@render header()}
	</div>

	{@render children?.()}

	{#if contents}
		<div class={cn("px-4 flex-1", classes.content)}>
			{@render contents()}
		</div>
	{/if}

	{#if actions}
		<div class={cn("py-2 px-1", classes.actions)}>
			{@render actions()}
		</div>
	{/if}
</div>
