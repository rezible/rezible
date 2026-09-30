<script lang="ts">
	import type { Snippet } from "svelte";
	import { cn } from "$lib/utils";

	type Props = {
		/** Element id for aria-labelledby on the parent section. */
		id: string;
		title: string;
		/** Shown muted after the title. Omit when unknown. */
		count?: number;
		level?: 2 | 3;
		/** Right-aligned compact actions. */
		actions?: Snippet;
		class?: string;
	};

	const { id, title, count, level = 2, actions, class: className }: Props = $props();

	const headingClass = $derived(
		level === 2 ? "text-lg leading-[26px] font-semibold" : "text-[15px] leading-[22px] font-semibold"
	);
</script>

<header class={cn("flex min-h-8 flex-wrap items-center justify-between gap-x-3 gap-y-1", className)}>
	<svelte:element this={`h${level}`} {id} class={cn("flex items-baseline gap-2", headingClass)}>
		{title}
		{#if count !== undefined}
			<span class="text-sm font-normal text-muted-foreground tabular-nums">{count}</span>
		{/if}
	</svelte:element>
	{#if actions}
		<div class="flex shrink-0 items-center gap-2">
			{@render actions()}
		</div>
	{/if}
</header>
