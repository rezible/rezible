<script lang="ts">
	import { cn } from "$lib/utils";
	import type { Snippet } from "svelte";

	type Classes = {
		root?: string;
		container?: string;
		title?: string;
		subheading?: string;
	};
	type Props = {
		title: string | Snippet;
		subheading?: string | Snippet;
		avatar?: Snippet;
		actions?: Snippet;
		classes?: Classes;
	};
	const { title, subheading, avatar, actions, classes = {} }: Props = $props();
</script>

<div class={cn("flex items-center gap-4", classes.root)}>
	{@render avatar?.()}

	<div class={cn("flex-1", classes.container)}>
		{#if typeof title === "string"}
			<div class={cn("text-lg", classes.title)}>{title}</div>
		{:else}
			{@render title?.()}
		{/if}

		{#if typeof subheading === "string"}
			<div class={cn("text-sm text-surface-content/50", classes.subheading)}>{subheading}</div>
		{:else}
			{@render subheading?.()}
		{/if}
	</div>

	{@render actions?.()}
</div>
